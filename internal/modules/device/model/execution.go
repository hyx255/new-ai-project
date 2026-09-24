package model

import (
	"encoding/json"
	"time"
)

// ExecutionStatus represents the lifecycle state of an Execution.
type ExecutionStatus string

const (
	ExecutionStatusPending  ExecutionStatus = "PENDING"
	ExecutionStatusRunning  ExecutionStatus = "RUNNING"
	ExecutionStatusSuccess  ExecutionStatus = "SUCCESS"
	ExecutionStatusFailed   ExecutionStatus = "FAILED"
	ExecutionStatusRetrying ExecutionStatus = "RETRYING"
	ExecutionStatusSkipped  ExecutionStatus = "SKIPPED"
)

// IsValid checks if the status is valid.
func (s ExecutionStatus) IsValid() bool {
	switch s {
	case ExecutionStatusPending, ExecutionStatusRunning, ExecutionStatusSuccess,
		ExecutionStatusFailed, ExecutionStatusRetrying, ExecutionStatusSkipped:
		return true
	}
	return false
}

// IsTerminal returns true if the status is a final state.
func (s ExecutionStatus) IsTerminal() bool {
	return s == ExecutionStatusSuccess || s == ExecutionStatusFailed || s == ExecutionStatusSkipped
}

// IsActive returns true if the execution is currently running or retrying.
func (s ExecutionStatus) IsActive() bool {
	return s == ExecutionStatusRunning || s == ExecutionStatusRetrying
}

// Execution represents a single device execution unit.
type Execution struct {
	ID           string          `json:"id"`
	OperationID  string          `json:"operation_id"`
	DeviceID     string          `json:"device_id"`
	Status       ExecutionStatus `json:"status"`
	RetryCount   int             `json:"retry_count"`
	MaxRetries   int             `json:"max_retries"`
	Request      json.RawMessage `json:"request,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	Error        string          `json:"error,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// NewExecution creates a new Execution with PENDING status.
func NewExecution(id, operationID, deviceID string, maxRetries int) *Execution {
	now := time.Now()
	return &Execution{
		ID:          id,
		OperationID: operationID,
		DeviceID:    deviceID,
		Status:      ExecutionStatusPending,
		RetryCount:  0,
		MaxRetries:  maxRetries,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// TransitionRunning marks the execution as running.
func (e *Execution) TransitionRunning() {
	e.Status = ExecutionStatusRunning
	now := time.Now()
	if e.StartedAt == nil {
		e.StartedAt = &now
	}
	e.UpdatedAt = now
}

// TransitionSuccess marks the execution as successful.
func (e *Execution) TransitionSuccess(result json.RawMessage) {
	e.Status = ExecutionStatusSuccess
	e.Result = result
	now := time.Now()
	e.FinishedAt = &now
	e.UpdatedAt = now
}

// TransitionFailed marks the execution as failed with an error.
func (e *Execution) TransitionFailed(err string) {
	e.Status = ExecutionStatusFailed
	e.Error = err
	now := time.Now()
	e.FinishedAt = &now
	e.UpdatedAt = now
}

// TransitionRetrying marks the execution as retrying.
func (e *Execution) TransitionRetrying() {
	e.Status = ExecutionStatusRetrying
	e.RetryCount++
	e.UpdatedAt = time.Now()
}

// TransitionSkipped marks the execution as skipped (precondition not met).
func (e *Execution) TransitionSkipped(reason string) {
	e.Status = ExecutionStatusSkipped
	e.Error = reason
	now := time.Now()
	e.FinishedAt = &now
	e.UpdatedAt = now
}

// CanRetry returns true if the execution can be retried.
func (e *Execution) CanRetry() bool {
	return e.Status == ExecutionStatusFailed && e.RetryCount < e.MaxRetries
}