package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/platform/database"
)

// DeviceListParams represents query parameters for listing devices.
type DeviceListParams struct {
	Page     int
	PageSize int
	Search   string
}

// DeviceListResult represents a paginated list of devices.
type DeviceListResult struct {
	Items    []*model.Device
	Total    int
	Page     int
	PageSize int
}

// DeviceRepository defines the interface for Device persistence.
type DeviceRepository interface {
	Create(ctx context.Context, d *model.Device) error
	GetByID(ctx context.Context, id string) (*model.Device, error)
	List(ctx context.Context, params DeviceListParams) (*DeviceListResult, error)
	Update(ctx context.Context, d *model.Device) error
	UpdateStatus(ctx context.Context, id string, status model.DeviceStatus) error
}

// SQLDeviceRepository implements DeviceRepository using SQL.
type SQLDeviceRepository struct {
	db *database.DB
}

// NewDeviceRepository creates a new SQL-based Device repository.
func NewDeviceRepository(db *database.DB) DeviceRepository {
	return &SQLDeviceRepository{db: db}
}

func (r *SQLDeviceRepository) Create(ctx context.Context, d *model.Device) error {
	var lastOnline *time.Time
	if d.LastOnlineAt != nil {
		lastOnline = d.LastOnlineAt
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO devices (id, name, device_type_id, address, status, last_online_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Name, d.DeviceTypeID, d.Address, string(d.Status),
		lastOnline, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert device: %w", err)
	}
	return nil
}

func (r *SQLDeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	d := &model.Device{}
	var status string
	var lastOnline sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, device_type_id, address, status, last_online_at, created_at, updated_at
		 FROM devices WHERE id = ?`, id,
	).Scan(&d.ID, &d.Name, &d.DeviceTypeID, &d.Address, &status, &lastOnline,
		&d.CreatedAt, &d.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query device: %w", err)
	}

	d.Status = model.DeviceStatus(status)
	if lastOnline.Valid {
		t := lastOnline.Time
		d.LastOnlineAt = &t
	}

	return d, nil
}

func (r *SQLDeviceRepository) List(ctx context.Context, params DeviceListParams) (*DeviceListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize

	var total int
	var countErr error

	if params.Search != "" {
		countErr = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM devices WHERE name LIKE ?`,
			"%"+params.Search+"%",
		).Scan(&total)
	} else {
		countErr = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM devices`,
		).Scan(&total)
	}
	if countErr != nil {
		return nil, fmt.Errorf("count devices: %w", countErr)
	}

	var query string
	var args []interface{}

	if params.Search != "" {
		query = `SELECT id, name, device_type_id, address, status, last_online_at, created_at, updated_at
		         FROM devices WHERE name LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{"%" + params.Search + "%", params.PageSize, offset}
	} else {
		query = `SELECT id, name, device_type_id, address, status, last_online_at, created_at, updated_at
		         FROM devices ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = []interface{}{params.PageSize, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query devices: %w", err)
	}
	defer rows.Close()

	var items []*model.Device
	for rows.Next() {
		d := &model.Device{}
		var status string
		var lastOnline sql.NullTime

		if err := rows.Scan(&d.ID, &d.Name, &d.DeviceTypeID, &d.Address,
			&status, &lastOnline, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}

		d.Status = model.DeviceStatus(status)
		if lastOnline.Valid {
			t := lastOnline.Time
			d.LastOnlineAt = &t
		}

		items = append(items, d)
	}

	return &DeviceListResult{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, rows.Err()
}

func (r *SQLDeviceRepository) Update(ctx context.Context, d *model.Device) error {
	d.UpdatedAt = time.Now()

	var lastOnline *time.Time
	if d.LastOnlineAt != nil {
		lastOnline = d.LastOnlineAt
	}

	result, err := r.db.ExecContext(ctx,
		`UPDATE devices SET name=?, device_type_id=?, address=?, status=?, last_online_at=?, updated_at=?
		 WHERE id=?`,
		d.Name, d.DeviceTypeID, d.Address, string(d.Status), lastOnline, d.UpdatedAt, d.ID,
	)
	if err != nil {
		return fmt.Errorf("update device: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update device %s: %w", d.ID, ErrNotFound)
	}
	return nil
}

func (r *SQLDeviceRepository) UpdateStatus(ctx context.Context, id string, status model.DeviceStatus) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE devices SET status=?, updated_at=? WHERE id=?`,
		string(status), time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update device status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("update device status %s: %w", id, ErrNotFound)
	}
	return nil
}

