package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"broadcast-platform/internal/modules/device/gateway"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

// ExecutionTask represents a single device execution to be processed by the engine.
type ExecutionTask struct {
	Execution *model.Execution
	Device    *model.Device
	DeviceType *model.DeviceType
	OpType    model.OperationType
	Params    json.RawMessage
}

// ExecutionEngine processes device execution tasks asynchronously.
// It manages a worker pool to control concurrency and provides
// atomic progress tracking for batch operations.
type ExecutionEngine struct {
	workerPool    chan struct{} // semaphore for concurrency control
	operationRepo repository.OperationRepository
	executionRepo repository.ExecutionRepository
	deviceRepo    repository.DeviceRepository
	gateway       gateway.DeviceGateway
	logger        *logging.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewExecutionEngine creates a new execution engine.
// maxWorkers controls how many device executions can run concurrently.
func NewExecutionEngine(
	maxWorkers int,
	opRepo repository.OperationRepository,
	execRepo repository.ExecutionRepository,
	devRepo repository.DeviceRepository,
	gw gateway.DeviceGateway,
	logger *logging.Logger,
) *ExecutionEngine {
	if maxWorkers < 1 {
		maxWorkers = 10
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ExecutionEngine{
		workerPool:    make(chan struct{}, maxWorkers),
		operationRepo: opRepo,
		executionRepo: execRepo,
		deviceRepo:    devRepo,
		gateway:       gw,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Submit submits a batch of execution tasks for async processing.
// Returns immediately; tasks are executed in background goroutines.
// The operationID is used for progress tracking and final aggregation.
func (e *ExecutionEngine) Submit(operationID string, tasks []ExecutionTask) {
	if len(tasks) == 0 {
		// No valid tasks — operation should already be finalized by service
		return
	}

	for _, task := range tasks {
		e.wg.Add(1)
		go func(t ExecutionTask) {
			defer e.wg.Done()

			// Acquire worker slot (blocks if pool is full)
			select {
			case e.workerPool <- struct{}{}:
				defer func() { <-e.workerPool }()
			case <-e.ctx.Done():
				// Engine shutting down
				e.markExecutionFailed(t.Execution, "engine shutting down")
				e.incrementCounter(t.Execution.OperationID, "failed_count")
				e.tryFinalize(t.Execution.OperationID)
				return
			}

			e.processTask(t)
		}(task)
	}
}

// Shutdown stops the engine and waits for in-flight operations to complete.
func (e *ExecutionEngine) Shutdown(timeout time.Duration) {
	e.cancel()

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("execution engine: all operations completed")
	case <-time.After(timeout):
		e.logger.Warn("execution engine: shutdown timeout, some operations may be incomplete")
	}
}

// processTask handles a single execution task including mutex check, execution, and retry.
func (e *ExecutionEngine) processTask(task ExecutionTask) {
	exec := task.Execution
	opID := exec.OperationID
	device := task.Device
	opType := task.OpType

	// Check write operation mutex at execution time
	if opType.IsWriteOperation() {
		hasActive, err := e.executionRepo.HasActiveWriteExecution(e.ctx, device.ID, opID)
		if err != nil {
			e.logger.Error("engine: failed to check concurrent execution",
				zap.String("execution_id", exec.ID),
				zap.Error(err),
			)
		}
		if hasActive {
			exec.TransitionSkipped("device has an active write operation")
			e.persistExecution(exec)
			e.incrementCounter(opID, "skipped_count")
			e.tryFinalize(opID)
			return
		}
	}

	// Execute with retry
	e.executeWithRetry(exec, device, opType, task.Params)

	// Increment appropriate counter based on final status
	switch exec.Status {
	case model.ExecutionStatusSuccess:
		e.incrementCounter(opID, "success_count")
	case model.ExecutionStatusFailed:
		e.incrementCounter(opID, "failed_count")
	case model.ExecutionStatusSkipped:
		e.incrementCounter(opID, "skipped_count")
	}

	// Handle QUERY_STATUS device status update
	if opType == model.OperationTypeQueryStatus {
		e.handleQueryStatusResult(device, exec)
	}

	// Try to finalize operation
	e.tryFinalize(opID)
}

// executeWithRetry executes a gateway call with retry logic.
func (e *ExecutionEngine) executeWithRetry(exec *model.Execution, device *model.Device, opType model.OperationType, params json.RawMessage) {
	for {
		// Check engine context
		select {
		case <-e.ctx.Done():
			exec.TransitionFailed("engine shutting down")
			e.persistExecution(exec)
			return
		default:
		}

		// Transition to RUNNING
		exec.TransitionRunning()
		e.persistExecution(exec)

		// Execute gateway call
		result, err := e.executeGateway(device, opType, params)

		if err != nil {
			exec.Error = err.Error()
			exec.Result = nil

			// Check if can retry
			if exec.CanRetry() {
				exec.TransitionRetrying()
				e.persistExecution(exec)

				// Wait before retry (respect context)
				select {
				case <-time.After(opType.RetryInterval()):
					continue
				case <-e.ctx.Done():
					exec.TransitionFailed("engine shutting down during retry wait")
					e.persistExecution(exec)
					return
				}
			}

			// No more retries
			exec.TransitionFailed(err.Error())
			e.persistExecution(exec)
			return
		}

		// Success
		resultJSON, _ := json.Marshal(result)
		exec.TransitionSuccess(resultJSON)
		e.persistExecution(exec)
		return
	}
}

// executeGateway calls the appropriate gateway method based on operation type.
func (e *ExecutionEngine) executeGateway(device *model.Device, opType model.OperationType, params json.RawMessage) (interface{}, error) {
	switch opType {
	case model.OperationTypeQueryStatus:
		return e.gateway.QueryStatus(e.ctx, device)

	case model.OperationTypeSetVolume:
		var volParams struct {
			Volume int `json:"volume"`
		}
		if err := json.Unmarshal(params, &volParams); err != nil {
			return nil, fmt.Errorf("invalid parameters: %w", err)
		}
		return e.gateway.SetVolume(e.ctx, device, volParams.Volume)

	case model.OperationTypeRestart:
		restartResult, err := e.gateway.Restart(e.ctx, device)
		if err != nil {
			return nil, err
		}

		// Wait for device to come back online
		select {
		case <-time.After(2 * time.Second):
		case <-e.ctx.Done():
			return nil, errors.New("engine shutting down during restart wait")
		}

		// Verify device is back online
		statusResult, err := e.gateway.QueryStatus(e.ctx, device)
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
func (e *ExecutionEngine) handleQueryStatusResult(device *model.Device, exec *model.Execution) {
	var newStatus model.DeviceStatus

	if exec.Status == model.ExecutionStatusSuccess {
		newStatus = model.DeviceStatusActive
	} else {
		newStatus = model.DeviceStatusOffline
	}

	if device.Status != newStatus {
		if err := e.deviceRepo.UpdateStatus(e.ctx, device.ID, newStatus); err != nil {
			e.logger.Error("engine: failed to update device status after query",
				zap.String("device_id", device.ID),
				zap.Error(err),
			)
		} else {
			e.logger.Info("engine: device status updated after query",
				zap.String("device_id", device.ID),
				zap.String("old_status", string(device.Status)),
				zap.String("new_status", string(newStatus)),
			)
		}
	}
}

// persistExecution saves the current execution state to the database.
func (e *ExecutionEngine) persistExecution(exec *model.Execution) {
	if err := e.executionRepo.Update(e.ctx, exec); err != nil {
		e.logger.Error("engine: failed to persist execution",
			zap.String("execution_id", exec.ID),
			zap.String("status", string(exec.Status)),
			zap.Error(err),
		)
	}
}

// incrementCounter atomically increments an operation counter.
func (e *ExecutionEngine) incrementCounter(operationID string, field string) {
	if err := e.operationRepo.IncrementCounter(e.ctx, operationID, field); err != nil {
		e.logger.Error("engine: failed to increment counter",
			zap.String("operation_id", operationID),
			zap.String("field", field),
			zap.Error(err),
		)
	}
}

// tryFinalize checks if all executions are complete and finalizes the operation if so.
func (e *ExecutionEngine) tryFinalize(operationID string) {
	// Use AggregateAndFinalize with 0 deltas (counters already incremented)
	completed, op, err := e.operationRepo.AggregateAndFinalize(e.ctx, operationID, 0, 0, 0)
	if err != nil {
		e.logger.Error("engine: failed to check/finalize operation",
			zap.String("operation_id", operationID),
			zap.Error(err),
		)
		return
	}

	if completed {
		e.logger.Info("engine: operation finalized",
			zap.String("operation_id", operationID),
			zap.String("status", string(op.Status)),
			zap.Int("success", op.SuccessCount),
			zap.Int("failed", op.FailedCount),
			zap.Int("skipped", op.SkippedCount),
		)
	}
}

// markExecutionFailed marks an execution as failed without retry.
func (e *ExecutionEngine) markExecutionFailed(exec *model.Execution, reason string) {
	exec.TransitionFailed(reason)
	e.persistExecution(exec)
}