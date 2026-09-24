package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/platform/database"
)

// ExecutionRepository defines the interface for Execution persistence.
type ExecutionRepository interface {
	Create(ctx context.Context, exec *model.Execution) error
	GetByID(ctx context.Context, id string) (*model.Execution, error)
	GetByOperationID(ctx context.Context, operationID string) ([]*model.Execution, error)
	Update(ctx context.Context, exec *model.Execution) error
	HasActiveWriteExecution(ctx context.Context, deviceID string, excludeOperationID string) (bool, error)
}

type executionRepository struct {
	db *database.DB
}

func NewExecutionRepository(db *database.DB) ExecutionRepository {
	return &executionRepository{db: db}
}

func (r *executionRepository) Create(ctx context.Context, exec *model.Execution) error {
	var requestJSON, resultJSON []byte
	if exec.Request != nil {
		requestJSON = exec.Request
	}
	if exec.Result != nil {
		resultJSON = exec.Result
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO executions (id, operation_id, device_id, status, retry_count, max_retries,
		 request, result, error, started_at, finished_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		exec.ID, exec.OperationID, exec.DeviceID, string(exec.Status),
		exec.RetryCount, exec.MaxRetries, requestJSON, resultJSON, exec.Error,
		exec.StartedAt, exec.FinishedAt, exec.CreatedAt, exec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert execution: %w", err)
	}
	return nil
}

func (r *executionRepository) GetByID(ctx context.Context, id string) (*model.Execution, error) {
	exec := &model.Execution{}
	var requestJSON, resultJSON sql.NullString
	var statusStr string
	var errorStr sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, operation_id, device_id, status, retry_count, max_retries,
		 request, result, error, started_at, finished_at, created_at, updated_at
		 FROM executions WHERE id = ?`, id,
	).Scan(&exec.ID, &exec.OperationID, &exec.DeviceID, &statusStr,
		&exec.RetryCount, &exec.MaxRetries, &requestJSON, &resultJSON, &errorStr,
		&exec.StartedAt, &exec.FinishedAt, &exec.CreatedAt, &exec.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}

	exec.Status = model.ExecutionStatus(statusStr)

	if requestJSON.Valid && requestJSON.String != "" {
		exec.Request = json.RawMessage(requestJSON.String)
	}
	if resultJSON.Valid && resultJSON.String != "" {
		exec.Result = json.RawMessage(resultJSON.String)
	}
	if errorStr.Valid {
		exec.Error = errorStr.String
	}

	return exec, nil
}

func (r *executionRepository) GetByOperationID(ctx context.Context, operationID string) ([]*model.Execution, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, operation_id, device_id, status, retry_count, max_retries,
		 request, result, error, started_at, finished_at, created_at, updated_at
		 FROM executions WHERE operation_id = ?
		 ORDER BY created_at ASC`, operationID,
	)
	if err != nil {
		return nil, fmt.Errorf("query executions: %w", err)
	}
	defer rows.Close()

	var executions []*model.Execution
	for rows.Next() {
		exec := &model.Execution{}
		var requestJSON, resultJSON sql.NullString
		var statusStr string
		var errorStr sql.NullString

		if err := rows.Scan(&exec.ID, &exec.OperationID, &exec.DeviceID, &statusStr,
			&exec.RetryCount, &exec.MaxRetries, &requestJSON, &resultJSON, &errorStr,
			&exec.StartedAt, &exec.FinishedAt, &exec.CreatedAt, &exec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan execution: %w", err)
		}

		exec.Status = model.ExecutionStatus(statusStr)

		if requestJSON.Valid && requestJSON.String != "" {
			exec.Request = json.RawMessage(requestJSON.String)
		}
		if resultJSON.Valid && resultJSON.String != "" {
			exec.Result = json.RawMessage(resultJSON.String)
		}
		if errorStr.Valid {
			exec.Error = errorStr.String
		}

		executions = append(executions, exec)
	}

	return executions, rows.Err()
}

func (r *executionRepository) Update(ctx context.Context, exec *model.Execution) error {
	exec.UpdatedAt = time.Now()

	var requestJSON, resultJSON []byte
	if exec.Request != nil {
		requestJSON = exec.Request
	}
	if exec.Result != nil {
		resultJSON = exec.Result
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE executions SET status = ?, retry_count = ?, request = ?, result = ?,
		 error = ?, started_at = ?, finished_at = ?, updated_at = ?
		 WHERE id = ?`,
		string(exec.Status), exec.RetryCount, requestJSON, resultJSON,
		exec.Error, exec.StartedAt, exec.FinishedAt, exec.UpdatedAt, exec.ID,
	)
	if err != nil {
		return fmt.Errorf("update execution: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update execution %s: %w", exec.ID, ErrNotFound)
	}
	return nil
}

func (r *executionRepository) HasActiveWriteExecution(ctx context.Context, deviceID string, excludeOperationID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM executions
			  WHERE device_id = ? AND status IN (?, ?)
			  AND operation_id != ?`

	err := r.db.QueryRowContext(ctx, query,
		deviceID,
		string(model.ExecutionStatusRunning),
		string(model.ExecutionStatusRetrying),
		excludeOperationID,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("check active execution: %w", err)
	}

	return count > 0, nil
}