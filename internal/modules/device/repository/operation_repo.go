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

// OperationRepository defines the interface for Operation persistence.
type OperationRepository interface {
	Create(ctx context.Context, op *model.Operation) error
	GetByID(ctx context.Context, id string) (*model.Operation, error)
	ListByDeviceID(ctx context.Context, deviceID string, page, pageSize int) ([]*model.Operation, int, error)
	Update(ctx context.Context, op *model.Operation) error

	// Batch operation methods
	IncrementCounter(ctx context.Context, operationID string, field string) error
	AggregateAndFinalize(ctx context.Context, operationID string, successDelta, failedDelta, skippedDelta int) (bool, *model.Operation, error)
	UpdateStatus(ctx context.Context, operationID string, status model.OperationStatus) error
}

type operationRepository struct {
	db *database.DB
}

func NewOperationRepository(db *database.DB) OperationRepository {
	return &operationRepository{db: db}
}

func (r *operationRepository) Create(ctx context.Context, op *model.Operation) error {
	var paramsJSON []byte
	if op.Parameters != nil {
		paramsJSON = op.Parameters
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO operations (id, device_id, type, parameters, status,
		 total_count, success_count, failed_count, skipped_count, finished_at,
		 created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		op.ID, op.DeviceID, string(op.Type), paramsJSON, string(op.Status),
		op.TotalCount, op.SuccessCount, op.FailedCount, op.SkippedCount,
		op.FinishedAt, op.CreatedAt, op.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}
	return nil
}

func (r *operationRepository) GetByID(ctx context.Context, id string) (*model.Operation, error) {
	op := &model.Operation{}
	var paramsJSON sql.NullString
	var statusStr string
	var typeStr string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_id, type, parameters, status,
		 total_count, success_count, failed_count, skipped_count, finished_at,
		 created_at, updated_at
		 FROM operations WHERE id = ?`, id,
	).Scan(&op.ID, &op.DeviceID, &typeStr, &paramsJSON, &statusStr,
		&op.TotalCount, &op.SuccessCount, &op.FailedCount, &op.SkippedCount,
		&op.FinishedAt, &op.CreatedAt, &op.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query operation: %w", err)
	}

	op.Type = model.OperationType(typeStr)
	op.Status = model.OperationStatus(statusStr)

	if paramsJSON.Valid && paramsJSON.String != "" {
		op.Parameters = json.RawMessage(paramsJSON.String)
	}

	return op, nil
}

func (r *operationRepository) ListByDeviceID(ctx context.Context, deviceID string, page, pageSize int) ([]*model.Operation, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM operations WHERE device_id = ?`, deviceID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count operations: %w", err)
	}

	// Query page
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_id, type, parameters, status,
		 total_count, success_count, failed_count, skipped_count, finished_at,
		 created_at, updated_at
		 FROM operations WHERE device_id = ?
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		deviceID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query operations: %w", err)
	}
	defer rows.Close()

	var operations []*model.Operation
	for rows.Next() {
		op, err := scanOperation(rows)
		if err != nil {
			return nil, 0, err
		}
		operations = append(operations, op)
	}

	return operations, total, rows.Err()
}

func (r *operationRepository) Update(ctx context.Context, op *model.Operation) error {
	op.UpdatedAt = time.Now()

	var paramsJSON []byte
	if op.Parameters != nil {
		paramsJSON = op.Parameters
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE operations SET status = ?, parameters = ?,
		 total_count = ?, success_count = ?, failed_count = ?, skipped_count = ?,
		 finished_at = ?, updated_at = ?
		 WHERE id = ?`,
		string(op.Status), paramsJSON,
		op.TotalCount, op.SuccessCount, op.FailedCount, op.SkippedCount,
		op.FinishedAt, op.UpdatedAt, op.ID,
	)
	if err != nil {
		return fmt.Errorf("update operation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update operation %s: %w", op.ID, ErrNotFound)
	}
	return nil
}

func (r *operationRepository) UpdateStatus(ctx context.Context, operationID string, status model.OperationStatus) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE operations SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), time.Now(), operationID,
	)
	if err != nil {
		return fmt.Errorf("update operation status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update operation status %s: %w", operationID, ErrNotFound)
	}
	return nil
}

// IncrementCounter atomically increments one of the execution counters.
// field must be "success_count", "failed_count", or "skipped_count".
func (r *operationRepository) IncrementCounter(ctx context.Context, operationID string, field string) error {
	// Validate field name to prevent SQL injection
	switch field {
	case "success_count", "failed_count", "skipped_count":
		// valid
	default:
		return fmt.Errorf("invalid counter field: %s", field)
	}

	query := fmt.Sprintf(
		`UPDATE operations SET %s = %s + 1, updated_at = ? WHERE id = ?`,
		field, field,
	)

	_, err := r.db.ExecContext(ctx, query, time.Now(), operationID)
	if err != nil {
		return fmt.Errorf("increment counter %s: %w", field, err)
	}
	return nil
}

// AggregateAndFinalize atomically updates all counters and checks if the operation is complete.
// Returns (completed, updated operation, error).
func (r *operationRepository) AggregateAndFinalize(ctx context.Context, operationID string, successDelta, failedDelta, skippedDelta int) (bool, *model.Operation, error) {
	now := time.Now()

	// Atomic update of counters
	_, err := r.db.ExecContext(ctx,
		`UPDATE operations SET
		 success_count = success_count + ?,
		 failed_count = failed_count + ?,
		 skipped_count = skipped_count + ?,
		 updated_at = ?
		 WHERE id = ?`,
		successDelta, failedDelta, skippedDelta, now, operationID,
	)
	if err != nil {
		return false, nil, fmt.Errorf("update counters: %w", err)
	}

	// Read current state
	op, err := r.GetByID(ctx, operationID)
	if err != nil {
		return false, nil, fmt.Errorf("get operation after counter update: %w", err)
	}

	finished := (op.SuccessCount + op.FailedCount + op.SkippedCount) >= op.TotalCount

	if finished && !op.Status.IsTerminal() {
		op.AggregateFromCounters()
		_, err := r.db.ExecContext(ctx,
			`UPDATE operations SET status = ?, finished_at = ?, updated_at = ? WHERE id = ?`,
			string(op.Status), op.FinishedAt, op.UpdatedAt, operationID,
		)
		if err != nil {
			return false, nil, fmt.Errorf("finalize operation: %w", err)
		}
	}

	return finished, op, nil
}

// scanOperation scans a row into an Operation.
func scanOperation(rows *sql.Rows) (*model.Operation, error) {
	op := &model.Operation{}
	var paramsJSON sql.NullString
	var statusStr string
	var typeStr string

	err := rows.Scan(&op.ID, &op.DeviceID, &typeStr, &paramsJSON, &statusStr,
		&op.TotalCount, &op.SuccessCount, &op.FailedCount, &op.SkippedCount,
		&op.FinishedAt, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan operation: %w", err)
	}

	op.Type = model.OperationType(typeStr)
	op.Status = model.OperationStatus(statusStr)

	if paramsJSON.Valid && paramsJSON.String != "" {
		op.Parameters = json.RawMessage(paramsJSON.String)
	}

	return op, nil
}