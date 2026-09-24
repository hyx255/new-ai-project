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

// DeviceTypeRepository defines the interface for DeviceType persistence.
type DeviceTypeRepository interface {
	Create(ctx context.Context, dt *model.DeviceType) error
	GetByID(ctx context.Context, id string) (*model.DeviceType, error)
	List(ctx context.Context) ([]*model.DeviceType, error)
	Update(ctx context.Context, dt *model.DeviceType) error
	Delete(ctx context.Context, id string) error
	ExistsByNameVendorModel(ctx context.Context, name, vendor, model string, excludeID string) (bool, error)
	CountDevicesByTypeID(ctx context.Context, deviceTypeID string) (int, error)
}

// SQLDeviceTypeRepository implements DeviceTypeRepository using SQL.
type SQLDeviceTypeRepository struct {
	db *database.DB
}

// NewDeviceTypeRepository creates a new SQL-based DeviceType repository.
func NewDeviceTypeRepository(db *database.DB) DeviceTypeRepository {
	return &SQLDeviceTypeRepository{db: db}
}

func (r *SQLDeviceTypeRepository) Create(ctx context.Context, dt *model.DeviceType) error {
	capJSON, err := json.Marshal(dt.Capabilities)
	if err != nil {
		return fmt.Errorf("marshal capabilities: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO device_types (id, name, vendor, model, description, capabilities, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		dt.ID, dt.Name, dt.Vendor, dt.Model, dt.Description, string(capJSON),
		dt.CreatedAt, dt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert device type: %w", err)
	}
	return nil
}

func (r *SQLDeviceTypeRepository) GetByID(ctx context.Context, id string) (*model.DeviceType, error) {
	dt := &model.DeviceType{}
	var capJSON string

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, vendor, model, description, capabilities, created_at, updated_at
		 FROM device_types WHERE id = ?`, id,
	).Scan(&dt.ID, &dt.Name, &dt.Vendor, &dt.Model, &dt.Description, &capJSON,
		&dt.CreatedAt, &dt.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query device type: %w", err)
	}

	if capJSON != "" {
		if err := json.Unmarshal([]byte(capJSON), &dt.Capabilities); err != nil {
			return nil, fmt.Errorf("unmarshal capabilities: %w", err)
		}
	}

	return dt, nil
}

func (r *SQLDeviceTypeRepository) List(ctx context.Context) ([]*model.DeviceType, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, vendor, model, description, capabilities, created_at, updated_at
		 FROM device_types ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query device types: %w", err)
	}
	defer rows.Close()

	var result []*model.DeviceType
	for rows.Next() {
		dt := &model.DeviceType{}
		var capJSON string

		if err := rows.Scan(&dt.ID, &dt.Name, &dt.Vendor, &dt.Model,
			&dt.Description, &capJSON, &dt.CreatedAt, &dt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device type: %w", err)
		}

		if capJSON != "" {
			if err := json.Unmarshal([]byte(capJSON), &dt.Capabilities); err != nil {
				return nil, fmt.Errorf("unmarshal capabilities: %w", err)
			}
		}

		result = append(result, dt)
	}

	return result, rows.Err()
}

func (r *SQLDeviceTypeRepository) Update(ctx context.Context, dt *model.DeviceType) error {
	capJSON, err := json.Marshal(dt.Capabilities)
	if err != nil {
		return fmt.Errorf("marshal capabilities: %w", err)
	}

	dt.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx,
		`UPDATE device_types SET name=?, vendor=?, model=?, description=?, capabilities=?, updated_at=?
		 WHERE id=?`,
		dt.Name, dt.Vendor, dt.Model, dt.Description, string(capJSON), dt.UpdatedAt, dt.ID,
	)
	if err != nil {
		return fmt.Errorf("update device type: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update device type %s: %w", dt.ID, ErrNotFound)
	}
	return nil
}

func (r *SQLDeviceTypeRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM device_types WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete device type: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("delete device type %s: %w", id, ErrNotFound)
	}
	return nil
}

func (r *SQLDeviceTypeRepository) ExistsByNameVendorModel(ctx context.Context, name, vendor, model string, excludeID string) (bool, error) {
	var count int
	var err error

	if excludeID == "" {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM device_types WHERE name=? AND vendor=? AND model=?`,
			name, vendor, model,
		).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM device_types WHERE name=? AND vendor=? AND model=? AND id!=?`,
			name, vendor, model, excludeID,
		).Scan(&count)
	}

	if err != nil {
		return false, fmt.Errorf("check uniqueness: %w", err)
	}
	return count > 0, nil
}

func (r *SQLDeviceTypeRepository) CountDevicesByTypeID(ctx context.Context, deviceTypeID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM devices WHERE device_type_id=?`, deviceTypeID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count devices: %w", err)
	}
	return count, nil
}



