package service

import (
	"context"
	"errors"
	"fmt"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceTypeInvalid = errors.New("device type does not exist")
	ErrInvalidTransition = errors.New("invalid device status transition")
)

// DeviceService handles Device business logic.
type DeviceService struct {
	repo   repository.DeviceRepository
	dtRepo repository.DeviceTypeRepository
	logger *logging.Logger
}

// NewDeviceService creates a new DeviceService.
func NewDeviceService(
	repo repository.DeviceRepository,
	dtRepo repository.DeviceTypeRepository,
	logger *logging.Logger,
) *DeviceService {
	return &DeviceService{repo: repo, dtRepo: dtRepo, logger: logger}
}

// CreateDeviceParams represents input for creating a Device.
type CreateDeviceParams struct {
	Name         string
	DeviceTypeID string
	Address      string
}

// Create creates a new Device.
func (s *DeviceService) Create(ctx context.Context, params CreateDeviceParams) (*model.Device, error) {
	if params.Name == "" {
		return nil, fmt.Errorf("device name is required")
	}
	if params.DeviceTypeID == "" {
		return nil, fmt.Errorf("device type ID is required")
	}

	// Validate DeviceType exists (BR-002)
	_, err := s.dtRepo.GetByID(ctx, params.DeviceTypeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDeviceTypeInvalid
		}
		return nil, fmt.Errorf("check device type: %w", err)
	}

	id := model.GenerateDeviceID()
	device, err := model.NewDevice(id, params.Name, params.DeviceTypeID, params.Address)
	if err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("persist device: %w", err)
	}

	s.logger.Info("device created",
		zap.String("id", device.ID),
		zap.String("name", device.Name),
		zap.String("device_type_id", device.DeviceTypeID),
		zap.String("status", string(device.Status)),
	)

	return device, nil
}

// GetByID retrieves a Device by ID.
func (s *DeviceService) GetByID(ctx context.Context, id string) (*model.Device, error) {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	return device, nil
}

// GetByIDWithDeviceType retrieves a Device by ID along with its DeviceType.
func (s *DeviceService) GetByIDWithDeviceType(ctx context.Context, id string) (*model.Device, *model.DeviceType, error) {
	device, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	dt, err := s.dtRepo.GetByID(ctx, device.DeviceTypeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// DeviceType was deleted but Device still references it
			return device, nil, nil
		}
		return nil, nil, fmt.Errorf("get device type: %w", err)
	}

	return device, dt, nil
}

// List retrieves a paginated list of Devices.
func (s *DeviceService) List(ctx context.Context, page, pageSize int, search string) (*repository.DeviceListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return s.repo.List(ctx, repository.DeviceListParams{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	})
}

// UpdateDeviceParams represents input for updating a Device.
type UpdateDeviceParams struct {
	Name         string
	DeviceTypeID string
	Address      string
}

// Update updates an existing Device.
func (s *DeviceService) Update(ctx context.Context, id string, params UpdateDeviceParams) (*model.Device, error) {
	device, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.Name == "" {
		return nil, fmt.Errorf("device name is required")
	}

	// If DeviceTypeID is changing, validate new type exists
	if params.DeviceTypeID != "" && params.DeviceTypeID != device.DeviceTypeID {
		_, err := s.dtRepo.GetByID(ctx, params.DeviceTypeID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrDeviceTypeInvalid
			}
			return nil, fmt.Errorf("check device type: %w", err)
		}
		device.DeviceTypeID = params.DeviceTypeID
	}

	device.Name = params.Name
	device.Address = params.Address

	if err := device.Validate(); err != nil {
		return nil, fmt.Errorf("validate device: %w", err)
	}

	if err := s.repo.Update(ctx, device); err != nil {
		return nil, fmt.Errorf("update device: %w", err)
	}

	s.logger.Info("device updated",
		zap.String("id", device.ID),
		zap.String("name", device.Name),
	)

	return device, nil
}

// Disable transitions a Device to DISABLED status.
func (s *DeviceService) Disable(ctx context.Context, id string) (*model.Device, error) {
	device, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := device.Disable(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTransition, err)
	}

	if err := s.repo.UpdateStatus(ctx, device.ID, device.Status); err != nil {
		return nil, fmt.Errorf("update device status: %w", err)
	}

	s.logger.Info("device disabled",
		zap.String("id", device.ID),
		zap.String("status", string(device.Status)),
	)

	return device, nil
}

// Enable transitions a Device from DISABLED back to REGISTERED.
func (s *DeviceService) Enable(ctx context.Context, id string) (*model.Device, error) {
	device, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := device.Enable(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTransition, err)
	}

	if err := s.repo.UpdateStatus(ctx, device.ID, device.Status); err != nil {
		return nil, fmt.Errorf("update device status: %w", err)
	}

	s.logger.Info("device enabled",
		zap.String("id", device.ID),
		zap.String("status", string(device.Status)),
	)

	return device, nil
}
