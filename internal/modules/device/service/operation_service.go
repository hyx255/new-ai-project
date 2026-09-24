package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"broadcast-platform/internal/modules/device/gateway"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

var (
	ErrOperationNotFound    = errors.New("operation not found")
	ErrDeviceDisabled       = errors.New("device is disabled")
	ErrDeviceOffline        = errors.New("device is offline")
	ErrCapabilityRequired   = errors.New("device does not support required capability")
	ErrInvalidParameter     = errors.New("invalid parameter")
	ErrConcurrentOperation  = errors.New("device has an active operation")
)

// SetVolumeParams represents parameters for SET_VOLUME operation.
type SetVolumeParams struct {
	Volume int `json:"volume"`
}

// OperationService handles device operation business logic.
type OperationService struct {
	deviceRepo   repository.DeviceRepository
	deviceTypeRepo repository.DeviceTypeRepository
	operationRepo repository.OperationRepository
	executionRepo repository.ExecutionRepository
	gateway      gateway.DeviceGateway
	logger       *logging.Logger
}

// NewOperationService creates a new OperationService.
func NewOperationService(
	deviceRepo repository.DeviceRepository,
	deviceTypeRepo repository.DeviceTypeRepository,
	operationRepo repository.OperationRepository,
	executionRepo repository.ExecutionRepository,
	gw gateway.DeviceGateway,
	logger *logging.Logger,
) *OperationService {
	return &OperationService{
		deviceRepo:    deviceRepo,
		deviceTypeRepo: deviceTypeRepo,
		operationRepo: operationRepo,
		executionRepo: executionRepo,
		gateway:       gw,
		logger:        logger,
	}
}

// CreateOperationParams represents input for creating an Operation.
type CreateOperationParams struct {
	DeviceID   string
	Type       model.OperationType
	Parameters json.RawMessage
}

// CreateOperationResult represents the result of creating an Operation.
type CreateOperationResult struct {
	Operation *model.Operation
	Execution *model.Execution
}

// Create creates and executes a device operation.
// Flow: Create Operation → Create Execution(PENDING) → Execute with retries → Update states
func (s *OperationService) Create(ctx context.Context, params CreateOperationParams) (*CreateOperationResult, error) {
	// 1. Validate device exists and get DeviceType
	device, err := s.deviceRepo.GetByID(ctx, params.DeviceID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("get device: %w", err)
	}

	deviceType, err := s.deviceTypeRepo.GetByID(ctx, device.DeviceTypeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDeviceTypeInvalid
		}
		return nil, fmt.Errorf("get device type: %w", err)
	}

	// 2. Check device status allows operation
	if device.Status == model.DeviceStatusDisabled {
		return nil, ErrDeviceDisabled
	}

	// 3. Capability check BEFORE gateway call
	requiredCap := params.Type.ToCapability()
	if !deviceType.HasCapability(requiredCap) {
		return nil, fmt.Errorf("%w: %s", ErrCapabilityRequired, requiredCap)
	}

	// 4. Validate parameters
	if err := s.validateParameters(params.Type, params.Parameters); err != nil {
		return nil, err
	}

	// 5. Check concurrent write operations
	if params.Type.IsWriteOperation() {
		tempOpID := model.GenerateOperationID()
		hasActive, err := s.executionRepo.HasActiveWriteExecution(ctx, params.DeviceID, tempOpID)
		if err != nil {
			return nil, fmt.Errorf("check concurrent operation: %w", err)
		}
		if hasActive {
			return nil, ErrConcurrentOperation
		}
	}

	// 6. Create Operation (PENDING)
	opID := model.GenerateOperationID()
	operation := model.NewOperation(opID, params.DeviceID, params.Type, params.Parameters)

	if err := s.operationRepo.Create(ctx, operation); err != nil {
		return nil, fmt.Errorf("create operation: %w", err)
	}

	// 7. Create Execution (PENDING) - MUST persist before execution
	execID := model.GenerateExecutionID()
	execution := model.NewExecution(execID, opID, params.DeviceID, params.Type.MaxRetries())

	if err := s.executionRepo.Create(ctx, execution); err != nil {
		return nil, fmt.Errorf("create execution: %w", err)
	}

	// 8. Transition Operation to IN_PROGRESS
	operation.TransitionInProgress()
	if err := s.operationRepo.Update(ctx, operation); err != nil {
		return nil, fmt.Errorf("update operation: %w", err)
	}

	// 9. Execute with retry logic
	s.executeWithRetry(ctx, execution, device, params.Type, params.Parameters)

	// 10. Aggregate Operation status from Execution
	operation.AggregateStatusFromExecution(execution.Status)
	if err := s.operationRepo.Update(ctx, operation); err != nil {
		s.logger.Error("failed to update operation status", zap.Error(err))
	}

	// 11. Handle QUERY_STATUS special case: update device status
	if params.Type == model.OperationTypeQueryStatus {
		s.handleQueryStatusResult(ctx, device, execution)
	}

	s.logger.Info("operation completed",
		zap.String("operation_id", operation.ID),
		zap.String("type", string(operation.Type)),
		zap.String("operation_status", string(operation.Status)),
		zap.String("execution_status", string(execution.Status)),
	)

	return &CreateOperationResult{
		Operation: operation,
		Execution: execution,
	}, nil
}

// validateParameters validates operation-specific parameters.
func (s *OperationService) validateParameters(opType model.OperationType, params json.RawMessage) error {
	switch opType {
	case model.OperationTypeSetVolume:
		var volParams SetVolumeParams
		if err := json.Unmarshal(params, &volParams); err != nil {
			return fmt.Errorf("%w: invalid volume parameters", ErrInvalidParameter)
		}
		if volParams.Volume < 0 || volParams.Volume > 100 {
			return fmt.Errorf("%w: volume must be 0-100", ErrInvalidParameter)
		}
	case model.OperationTypeQueryStatus, model.OperationTypeRestart:
		// No parameters required
	}
	return nil
}

// executeWithRetry executes the gateway call with retry logic.
func (s *OperationService) executeWithRetry(ctx context.Context, exec *model.Execution, device *model.Device, opType model.OperationType, params json.RawMessage) {
	for {
		// Transition to RUNNING
		exec.TransitionRunning()
		if err := s.executionRepo.Update(ctx, exec); err != nil {
			s.logger.Error("failed to update execution to running", zap.Error(err))
		}

		// Execute gateway call
		result, err := s.executeGateway(ctx, device, opType, params)

		if err != nil {
			// Execution failed
			exec.Error = err.Error()
			exec.Result = nil

			// Check if can retry
			if exec.CanRetry() {
				exec.TransitionRetrying()
				if err := s.executionRepo.Update(ctx, exec); err != nil {
					s.logger.Error("failed to update execution to retrying", zap.Error(err))
				}

				// Wait before retry
				time.Sleep(opType.RetryInterval())
				continue
			}

			// No more retries, mark as FAILED
			exec.TransitionFailed(err.Error())
			if err := s.executionRepo.Update(ctx, exec); err != nil {
				s.logger.Error("failed to update execution to failed", zap.Error(err))
			}
			return
		}

		// Success
		resultJSON, _ := json.Marshal(result)
		exec.TransitionSuccess(resultJSON)
		if err := s.executionRepo.Update(ctx, exec); err != nil {
			s.logger.Error("failed to update execution to success", zap.Error(err))
		}
		return
	}
}

// executeGateway calls the appropriate gateway method based on operation type.
func (s *OperationService) executeGateway(ctx context.Context, device *model.Device, opType model.OperationType, params json.RawMessage) (interface{}, error) {
	switch opType {
	case model.OperationTypeQueryStatus:
		return s.gateway.QueryStatus(ctx, device)

	case model.OperationTypeSetVolume:
		var volParams SetVolumeParams
		if err := json.Unmarshal(params, &volParams); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
		return s.gateway.SetVolume(ctx, device, volParams.Volume)

	case model.OperationTypeRestart:
		// RESTART requires special handling: send restart, wait for recovery, verify with QueryStatus
		restartResult, err := s.gateway.Restart(ctx, device)
		if err != nil {
			return nil, err
		}

		// Wait for device to come back online (2 seconds)
		time.Sleep(2 * time.Second)

		// Verify device is back online
		statusResult, err := s.gateway.QueryStatus(ctx, device)
		if err != nil {
			return nil, fmt.Errorf("restart verification failed: %w", err)
		}

		if !statusResult.Online {
			return nil, errors.New("device did not come back online after restart")
		}

		return restartResult, nil

	default:
		return nil, fmt.Errorf("unsupported operation type: %s", opType)
	}
}

// handleQueryStatusResult updates device status based on QUERY_STATUS result.
func (s *OperationService) handleQueryStatusResult(ctx context.Context, device *model.Device, exec *model.Execution) {
	var newStatus model.DeviceStatus

	if exec.Status == model.ExecutionStatusSuccess {
		newStatus = model.DeviceStatusActive
	} else {
		// Failed/timeout/offline error → OFFLINE
		newStatus = model.DeviceStatusOffline
	}

	// Only update if status changed
	if device.Status != newStatus {
		if err := s.deviceRepo.UpdateStatus(ctx, device.ID, newStatus); err != nil {
			s.logger.Error("failed to update device status after query",
				zap.String("device_id", device.ID),
				zap.String("new_status", string(newStatus)),
				zap.Error(err),
			)
		} else {
			s.logger.Info("device status updated after query",
				zap.String("device_id", device.ID),
				zap.String("old_status", string(device.Status)),
				zap.String("new_status", string(newStatus)),
			)
		}
	}
}

// GetOperation retrieves an Operation by ID.
func (s *OperationService) GetOperation(ctx context.Context, id string) (*model.Operation, error) {
	op, err := s.operationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrOperationNotFound
		}
		return nil, fmt.Errorf("get operation: %w", err)
	}
	return op, nil
}

// GetOperationWithExecutions retrieves an Operation and its Executions.
func (s *OperationService) GetOperationWithExecutions(ctx context.Context, id string) (*model.Operation, []*model.Execution, error) {
	op, err := s.GetOperation(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	execs, err := s.executionRepo.GetByOperationID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get executions: %w", err)
	}

	return op, execs, nil
}

// ListDeviceOperations retrieves paginated Operations for a device.
func (s *OperationService) ListDeviceOperations(ctx context.Context, deviceID string, page, pageSize int) ([]*model.Operation, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	operations, total, err := s.operationRepo.ListByDeviceID(ctx, deviceID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list operations: %w", err)
	}

	return operations, total, nil
}