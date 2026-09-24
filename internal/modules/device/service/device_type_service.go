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
	ErrDeviceTypeNotFound      = errors.New("device type not found")
	ErrDeviceTypeAlreadyExists = errors.New("device type already exists (name+vendor+model)")
	ErrDeviceTypeInUse         = errors.New("device type has associated devices, cannot delete")
	ErrCannotRemoveCapability  = errors.New("cannot remove capability that is in use")
)

// DeviceTypeService handles DeviceType business logic.
type DeviceTypeService struct {
	repo   repository.DeviceTypeRepository
	logger *logging.Logger
}

// NewDeviceTypeService creates a new DeviceTypeService.
func NewDeviceTypeService(repo repository.DeviceTypeRepository, logger *logging.Logger) *DeviceTypeService {
	return &DeviceTypeService{repo: repo, logger: logger}
}

// CreateDeviceTypeParams represents input for creating a DeviceType.
type CreateDeviceTypeParams struct {
	Name         string            
	Vendor       string            
	Model        string            
	Description  string            
	Capabilities []model.Capability
}

// Create creates a new DeviceType.
func (s *DeviceTypeService) Create(ctx context.Context, params CreateDeviceTypeParams) (*model.DeviceType, error) {
	if params.Name == "" || params.Vendor == "" || params.Model == "" {
		return nil, fmt.Errorf("name, vendor, and model are required")
	}

	for _, cap := range params.Capabilities {
		if !cap.IsValid() {
			return nil, fmt.Errorf("invalid capability: %s", cap)
		}
	}

	exists, err := s.repo.ExistsByNameVendorModel(ctx, params.Name, params.Vendor, params.Model, "")
	if err != nil {
		return nil, fmt.Errorf("check uniqueness: %w", err)
	}
	if exists {
		return nil, ErrDeviceTypeAlreadyExists
	}

	id := model.GenerateDeviceTypeID()
	dt, err := model.NewDeviceType(id, params.Name, params.Vendor, params.Model, params.Description, params.Capabilities)
	if err != nil {
		return nil, fmt.Errorf("create device type: %w", err)
	}

	if err := s.repo.Create(ctx, dt); err != nil {
		return nil, fmt.Errorf("persist device type: %w", err)
	}

	s.logger.Info("device type created",
		zap.String("id", dt.ID),
		zap.String("name", dt.Name),
		zap.String("vendor", dt.Vendor),
		zap.String("model", dt.Model),
	)

	return dt, nil
}

// GetByID retrieves a DeviceType by ID.
func (s *DeviceTypeService) GetByID(ctx context.Context, id string) (*model.DeviceType, error) {
	dt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrDeviceTypeNotFound
		}
		return nil, fmt.Errorf("get device type: %w", err)
	}
	return dt, nil
}

// List retrieves all DeviceTypes.
func (s *DeviceTypeService) List(ctx context.Context) ([]*model.DeviceType, error) {
	return s.repo.List(ctx)
}

// UpdateDeviceTypeParams represents input for updating a DeviceType.
type UpdateDeviceTypeParams struct {
	Name         string            
	Vendor       string            
	Model        string            
	Description  string            
	Capabilities []model.Capability
}

// Update updates an existing DeviceType.
func (s *DeviceTypeService) Update(ctx context.Context, id string, params UpdateDeviceTypeParams) (*model.DeviceType, error) {
	dt, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.Name == "" || params.Vendor == "" || params.Model == "" {
		return nil, fmt.Errorf("name, vendor, and model are required")
	}

	for _, cap := range params.Capabilities {
		if !cap.IsValid() {
			return nil, fmt.Errorf("invalid capability: %s", cap)
		}
	}

	if params.Name != dt.Name || params.Vendor != dt.Vendor || params.Model != dt.Model {
		exists, err := s.repo.ExistsByNameVendorModel(ctx, params.Name, params.Vendor, params.Model, id)
		if err != nil {
			return nil, fmt.Errorf("check uniqueness: %w", err)
		}
		if exists {
			return nil, ErrDeviceTypeAlreadyExists
		}
	}

	// Check if capabilities are being removed when devices exist
	deviceCount, err := s.repo.CountDevicesByTypeID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("count devices: %w", err)
	}
	if deviceCount > 0 {
		for _, existingCap := range dt.Capabilities {
			found := false
			for _, newCap := range params.Capabilities {
				if newCap == existingCap {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("%w: %s", ErrCannotRemoveCapability, existingCap)
			}
		}
	}

	dt.Name = params.Name
	dt.Vendor = params.Vendor
	dt.Model = params.Model
	dt.Description = params.Description
	dt.Capabilities = params.Capabilities

	if err := dt.Validate(); err != nil {
		return nil, fmt.Errorf("validate device type: %w", err)
	}

	if err := s.repo.Update(ctx, dt); err != nil {
		return nil, fmt.Errorf("update device type: %w", err)
	}

	s.logger.Info("device type updated",
		zap.String("id", dt.ID),
		zap.String("name", dt.Name),
	)

	return dt, nil
}

// Delete deletes a DeviceType if no devices reference it.
func (s *DeviceTypeService) Delete(ctx context.Context, id string) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	deviceCount, err := s.repo.CountDevicesByTypeID(ctx, id)
	if err != nil {
		return fmt.Errorf("count devices: %w", err)
	}
	if deviceCount > 0 {
		return ErrDeviceTypeInUse
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete device type: %w", err)
	}

	s.logger.Info("device type deleted", zap.String("id", id))
	return nil
}

