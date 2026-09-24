package model

import (
	"encoding/json"
	"time"
)

// OperationStatus represents the lifecycle state of an Operation.
type OperationStatus string

const (
	OperationStatusPending        OperationStatus = "PENDING"
	OperationStatusInProgress     OperationStatus = "IN_PROGRESS"
	OperationStatusCompleted      OperationStatus = "COMPLETED"
	OperationStatusFailed         OperationStatus = "FAILED"
	OperationStatusPartialSuccess OperationStatus = "PARTIAL_SUCCESS"
)

// IsValid checks if the status is valid.
func (s OperationStatus) IsValid() bool {
	switch s {
	case OperationStatusPending, OperationStatusInProgress,
		OperationStatusCompleted, OperationStatusFailed,
		OperationStatusPartialSuccess:
		return true
	}
	return false
}

// IsTerminal returns true if the status is a final state.
func (s OperationStatus) IsTerminal() bool {
	return s == OperationStatusCompleted ||
		s == OperationStatusFailed ||
		s == OperationStatusPartialSuccess
}

// Operation represents a business-level device operation intent.
// Single-device: DeviceID is set, TotalCount = 1.
// Batch: DeviceID is "", TotalCount = N.
type Operation struct {
	ID           string          `json:"id"`
	DeviceID     string          `json:"device_id,omitempty"`
	Type         OperationType   `json:"type"`
	Parameters   json.RawMessage `json:"parameters,omitempty"`
	Status       OperationStatus `json:"status"`

	// Batch counters (also maintained for single-device ops)
	TotalCount   int  `json:"total_count"`
	SuccessCount int  `json:"success_count"`
	FailedCount  int  `json:"failed_count"`
	SkippedCount int  `json:"skipped_count"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// NewOperation creates a new single-device Operation with PENDING status.
func NewOperation(id, deviceID string, opType OperationType, parameters json.RawMessage) *Operation {
	now := time.Now()
	return &Operation{
		ID:         id,
		DeviceID:   deviceID,
		Type:       opType,
		Parameters: parameters,
		Status:     OperationStatusPending,
		TotalCount: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewBatchOperation creates a new batch Operation with PENDING status.
func NewBatchOperation(id string, opType OperationType, parameters json.RawMessage, totalCount int) *Operation {
	now := time.Now()
	return &Operation{
		ID:         id,
		DeviceID:   "",
		Type:       opType,
		Parameters: parameters,
		Status:     OperationStatusPending,
		TotalCount: totalCount,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsBatch returns true if this is a batch operation.
func (o *Operation) IsBatch() bool {
	return o.DeviceID == ""
}

// RunningCount returns the number of executions still in progress.
func (o *Operation) RunningCount() int {
	running := o.TotalCount - o.SuccessCount - o.FailedCount - o.SkippedCount
	if running < 0 {
		return 0
	}
	return running
}

// TransitionInProgress marks the operation as in progress.
func (o *Operation) TransitionInProgress() {
	o.Status = OperationStatusInProgress
	o.UpdatedAt = time.Now()
}

// TransitionCompleted marks the operation as completed.
func (o *Operation) TransitionCompleted() {
	o.Status = OperationStatusCompleted
	o.UpdatedAt = time.Now()
}

// TransitionFailed marks the operation as failed.
func (o *Operation) TransitionFailed() {
	o.Status = OperationStatusFailed
	o.UpdatedAt = time.Now()
}

// TransitionPartialSuccess marks the operation as partially successful.
func (o *Operation) TransitionPartialSuccess() {
	o.Status = OperationStatusPartialSuccess
	o.UpdatedAt = time.Now()
}

// AggregateStatusFromExecution derives Operation status from a single Execution's final status.
// Used by single-device synchronous operations only.
func (o *Operation) AggregateStatusFromExecution(execStatus ExecutionStatus) {
	now := time.Now()
	o.FinishedAt = &now

	switch execStatus {
	case ExecutionStatusSuccess:
		o.SuccessCount = 1
		o.TransitionCompleted()
	case ExecutionStatusFailed, ExecutionStatusSkipped:
		if execStatus == ExecutionStatusSkipped {
			o.SkippedCount = 1
		} else {
			o.FailedCount = 1
		}
		o.TransitionFailed()
	}
}

// AggregateFromCounters computes the final Operation status from current counters.
// Called by the engine when all executions are complete.
func (o *Operation) AggregateFromCounters() {
	now := time.Now()
	o.FinishedAt = &now

	switch {
	case o.SuccessCount == o.TotalCount:
		o.Status = OperationStatusCompleted
	case o.SuccessCount > 0:
		o.Status = OperationStatusPartialSuccess
	default:
		o.Status = OperationStatusFailed
	}
	o.UpdatedAt = now
}