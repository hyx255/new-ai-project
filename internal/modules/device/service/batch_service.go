package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"broadcast-platform/internal/modules/device/engine"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

var (
	ErrEmptyDeviceIDs    = errors.New("device_ids must not be empty")
	ErrBatchTooLarge     = errors.New("too many devices in batch")
	ErrDuplicateDeviceID = errors.New("duplicate device_id in batch")
)

// BatchOperationService handles batch (multi-device) operation logic.
type BatchOperationService struct {
	deviceRepo     repository.DeviceRepository
	deviceTypeRepo repository.DeviceTypeRepository
	operationRepo  repository.OperationRepository
	executionRepo  repository.ExecutionRepository
	engine         *engine.ExecutionEngine
	maxBatchSize   int
	logger         *logging.Logger
}

// NewBatchOperationService creates a new BatchOperationService.
func NewBatchOperationService(
	deviceRepo repository.DeviceRepository,
	deviceTypeRepo repository.DeviceTypeRepository,
	operationRepo repository.OperationRepository,
	executionRepo repository.ExecutionRepository,
	eng *engine.ExecutionEngine,
	maxBatchSize int,
	logger *logging.Logger,
) *BatchOperationService {
	if maxBatchSize < 1 {
		maxBatchSize = 100
	}
	return &BatchOperationService{
		deviceRepo:     deviceRepo,
		deviceTypeRepo: deviceTypeRepo,
		operationRepo:  operationRepo,
		executionRepo:  executionRepo,
		engine:         eng,
		maxBatchSize:   maxBatchSize,
		logger:         logger,
	}
}

// CreateBatchParams represents input for creating a batch Operation.
type CreateBatchParams struct {
	DeviceIDs  []string
	Type       model.OperationType
	Parameters json.RawMessage
}

// CreateBatchResult represents the result of creating a batch Operation.
type CreateBatchResult struct {
	Operation  *model.Operation
	Executions []*model.Execution
}

// CreateBatchOperation validates, creates Operation + Executions, and submits to engine.
// Flow:
//  1. Validate inputs
//  2. Fetch all devices and device types
//  3. Pre-check each device
//  4. Create Operation (PENDING)
//  5. Create Execution for each device (PENDING or SKIPPED)
//  6. Transition Operation to IN_PROGRESS
//  7. Submit valid tasks to engine (async)
//  8. Return immediately
func (s *BatchOperationService) CreateBatchOperation(ctx context.Context, params CreateBatchParams) (*CreateBatchResult, error) {
	// --- Validation ---
	if len(params.DeviceIDs) == 0 {
		return nil, ErrEmptyDeviceIDs
	}
	if len(params.DeviceIDs) > s.maxBatchSize {
		return nil, fmt.Errorf("%w: max %d devices", ErrBatchTooLarge, s.maxBatchSize)
	}
	if !params.Type.IsValid() {
		return nil, fmt.Errorf("invalid operation type: %s", params.Type)
	}

	// Check for duplicate device IDs
	seen := make(map[string]bool, len(params.DeviceIDs))
	for _, id := range params.DeviceIDs {
		if id == "" {
			return nil, fmt.Errorf("empty device_id in batch")
		}
		if seen[id] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateDeviceID, id)
		}
		seen[id] = true
	}

	// Validate parameters
	if err := validateOpParameters(params.Type, params.Parameters); err != nil {
		return nil, err
	}

	// --- Fetch all devices and device types ---
	type deviceInfo struct {
		device     *model.Device
		deviceType *model.DeviceType
	}

	deviceMap := make(map[string]deviceInfo, len(params.DeviceIDs))
	dtCache := make(map[string]*model.DeviceType) // cache device types by ID

	for _, deviceID := range params.DeviceIDs {
		device, err := s.deviceRepo.GetByID(ctx, deviceID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// Will be handled as SKIPPED below
				continue
			}
			return nil, fmt.Errorf("get device %s: %w", deviceID, err)
		}

		dt, ok := dtCache[device.DeviceTypeID]
		if !ok {
			var err error
			dt, err = s.deviceTypeRepo.GetByID(ctx, device.DeviceTypeID)
			if err != nil {
				return nil, fmt.Errorf("get device type %s: %w", device.DeviceTypeID, err)
			}
			dtCache[device.DeviceTypeID] = dt
		}

		deviceMap[deviceID] = deviceInfo{device: device, deviceType: dt}
	}

	// --- Create Operation ---
	opID := model.GenerateOperationID()
	operation := model.NewBatchOperation(opID, params.Type, params.Parameters, len(params.DeviceIDs))

	if err := s.operationRepo.Create(ctx, operation); err != nil {
		return nil, fmt.Errorf("create batch operation: %w", err)
	}

	// --- Create Executions with pre-checks ---
	requiredCap := params.Type.ToCapability()
	maxRetries := params.Type.MaxRetries()

	var allExecutions []*model.Execution
	var validTasks []engine.ExecutionTask
	skippedCount := 0

	for _, deviceID := range params.DeviceIDs {
		execID := model.GenerateExecutionID()
		exec := model.NewExecution(execID, opID, deviceID, maxRetries)

		info, found := deviceMap[deviceID]
		if !found {
			// Device not found → SKIPPED
			exec.TransitionSkipped("device not found")
			s.persistExecution(ctx, exec)
			skippedCount++
			allExecutions = append(allExecutions, exec)
			continue
		}

		device := info.device
		dt := info.deviceType

		// Pre-check: disabled
		if device.Status == model.DeviceStatusDisabled {
			exec.TransitionSkipped("device is disabled")
			s.persistExecution(ctx, exec)
			skippedCount++
			allExecutions = append(allExecutions, exec)
			continue
		}

		// Pre-check: capability
		if !dt.HasCapability(requiredCap) {
			exec.TransitionSkipped(fmt.Sprintf("capability not supported: %s", requiredCap))
			s.persistExecution(ctx, exec)
			skippedCount++
			allExecutions = append(allExecutions, exec)
			continue
		}

		// Pre-check: OFFLINE + write operation
		if device.Status == model.DeviceStatusOffline && params.Type.IsWriteOperation() {
			exec.TransitionSkipped("device is offline, write operation not allowed")
			s.persistExecution(ctx, exec)
			skippedCount++
			allExecutions = append(allExecutions, exec)
			continue
		}

		// Valid → PENDING, submit to engine
		s.persistExecution(ctx, exec)
		allExecutions = append(allExecutions, exec)

		validTasks = append(validTasks, engine.ExecutionTask{
			Execution:  exec,
			Device:     device,
			DeviceType: dt,
			OpType:     params.Type,
			Params:     params.Parameters,
		})
	}

	// --- Update operation with initial skipped count ---
	if skippedCount > 0 {
		for i := 0; i < skippedCount; i++ {
			if err := s.operationRepo.IncrementCounter(ctx, opID, "skipped_count"); err != nil {
				s.logger.Error("failed to increment skipped counter", zap.Error(err))
			}
		}
	}

	// --- Transition to IN_PROGRESS (or finalize if all skipped) ---
	if len(validTasks) == 0 {
		// All devices were pre-skipped → finalize immediately
		_, _, err := s.operationRepo.AggregateAndFinalize(ctx, opID, 0, 0, 0)
		if err != nil {
			s.logger.Error("failed to finalize all-skipped batch operation", zap.Error(err))
		}
	} else {
		// Transition to IN_PROGRESS
		if err := s.operationRepo.UpdateStatus(ctx, opID, model.OperationStatusInProgress); err != nil {
			return nil, fmt.Errorf("update operation to in_progress: %w", err)
		}

		// Submit to engine (async)
		s.engine.Submit(opID, validTasks)
	}

	// Re-read operation to get latest state
	operation, err := s.operationRepo.GetByID(ctx, opID)
	if err != nil {
		return nil, fmt.Errorf("re-read operation: %w", err)
	}

	return &CreateBatchResult{
		Operation:  operation,
		Executions: allExecutions,
	}, nil
}

// GetBatchOperation retrieves a batch Operation and all its Executions.
func (s *BatchOperationService) GetBatchOperation(ctx context.Context, id string) (*model.Operation, []*model.Execution, error) {
	op, err := s.operationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrOperationNotFound
		}
		return nil, nil, fmt.Errorf("get operation: %w", err)
	}

	execs, err := s.executionRepo.GetByOperationID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get executions: %w", err)
	}

	return op, execs, nil
}

// BatchProgress represents lightweight progress information for polling.
type BatchProgress struct {
	OperationID  string                `json:"operation_id"`
	Status       model.OperationStatus `json:"status"`
	TotalCount   int                   `json:"total_count"`
	SuccessCount int                   `json:"success_count"`
	FailedCount  int                   `json:"failed_count"`
	SkippedCount int                   `json:"skipped_count"`
	RunningCount int                   `json:"running_count"`
	IsComplete   bool                  `json:"is_complete"`
}

// GetProgress retrieves lightweight progress information for a batch Operation.
func (s *BatchOperationService) GetProgress(ctx context.Context, id string) (*BatchProgress, error) {
	op, err := s.operationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrOperationNotFound
		}
		return nil, fmt.Errorf("get operation: %w", err)
	}

	return &BatchProgress{
		OperationID:  op.ID,
		Status:       op.Status,
		TotalCount:   op.TotalCount,
		SuccessCount: op.SuccessCount,
		FailedCount:  op.FailedCount,
		SkippedCount: op.SkippedCount,
		RunningCount: op.RunningCount(),
		IsComplete:   op.Status.IsTerminal(),
	}, nil
}

// persistExecution saves an execution to the database.
func (s *BatchOperationService) persistExecution(ctx context.Context, exec *model.Execution) {
	if err := s.executionRepo.Create(ctx, exec); err != nil {
		s.logger.Error("failed to persist execution",
			zap.String("execution_id", exec.ID),
			zap.Error(err),
		)
	}
}

// validateOpParameters validates parameters for the given operation type.
func validateOpParameters(opType model.OperationType, params json.RawMessage) error {
	switch opType {
	case model.OperationTypeSetVolume:
		if params == nil {
			return fmt.Errorf("%w: volume parameter required", ErrInvalidParameter)
		}
		var volParams struct {
			Volume int `json:"volume"`
		}
		if err := json.Unmarshal(params, &volParams); err != nil {
			return fmt.Errorf("%w: invalid volume parameter", ErrInvalidParameter)
		}
		if volParams.Volume < 0 || volParams.Volume > 100 {
			return fmt.Errorf("%w: volume must be 0-100", ErrInvalidParameter)
		}
	case model.OperationTypeQueryStatus, model.OperationTypeRestart:
		// No required parameters
	}
	return nil
}