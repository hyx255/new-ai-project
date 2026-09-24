package model

import (
	"errors"
	"time"
)

var (
	ErrDeviceNameEmpty       = errors.New("device name is required")
	ErrDeviceTypeIDEmpty     = errors.New("device type ID is required")
	ErrInvalidDeviceStatus   = errors.New("invalid device status transition")
	ErrDeviceAlreadyDisabled = errors.New("device is already disabled")
	ErrDeviceAlreadyEnabled  = errors.New("device is already enabled")
)

// DeviceStatus represents the operational status of a device.
type DeviceStatus string

const (
	DeviceStatusRegistered DeviceStatus = "REGISTERED"
	DeviceStatusActive     DeviceStatus = "ACTIVE"
	DeviceStatusOffline    DeviceStatus = "OFFLINE"
	DeviceStatusDisabled   DeviceStatus = "DISABLED"
)

// IsValid checks if the status is valid.
func (s DeviceStatus) IsValid() bool {
	switch s {
	case DeviceStatusRegistered, DeviceStatusActive, DeviceStatusOffline, DeviceStatusDisabled:
		return true
	}
	return false
}

// IsActive returns true if the device is in a state that allows operations.
func (s DeviceStatus) IsActive() bool {
	return s == DeviceStatusActive
}

// IsDisabled returns true if the device is disabled.
func (s DeviceStatus) IsDisabled() bool {
	return s == DeviceStatusDisabled
}

// CanBeDisabled returns true if the device can be disabled from current status.
func (s DeviceStatus) CanBeDisabled() bool {
	switch s {
	case DeviceStatusRegistered, DeviceStatusActive, DeviceStatusOffline:
		return true
	}
	return false
}

// CanBeEnabled returns true if the device can be enabled from current status.
func (s DeviceStatus) CanBeEnabled() bool {
	return s == DeviceStatusDisabled
}

// Device represents a physical device instance.
type Device struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	DeviceTypeID string       `json:"device_type_id"`
	Address      string       `json:"address"`
	Status       DeviceStatus `json:"status"`
	LastOnlineAt *time.Time   `json:"last_online_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// NewDevice creates a new Device with validation.
func NewDevice(id, name, deviceTypeID, address string) (*Device, error) {
	d := &Device{
		ID:           id,
		Name:         name,
		DeviceTypeID: deviceTypeID,
		Address:      address,
		Status:       DeviceStatusRegistered,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return d, nil
}

// Validate checks Device business rules.
func (d *Device) Validate() error {
	if d.Name == "" {
		return ErrDeviceNameEmpty
	}
	if d.DeviceTypeID == "" {
		return ErrDeviceTypeIDEmpty
	}
	return nil
}

// Disable transitions the device to DISABLED status.
// Allowed from: REGISTERED, ACTIVE, OFFLINE
func (d *Device) Disable() error {
	if d.Status.IsDisabled() {
		return ErrDeviceAlreadyDisabled
	}
	if !d.Status.CanBeDisabled() {
		return ErrInvalidDeviceStatus
	}
	d.Status = DeviceStatusDisabled
	d.UpdatedAt = time.Now()
	return nil
}

// Enable transitions the device from DISABLED back to REGISTERED.
func (d *Device) Enable() error {
	if !d.Status.IsDisabled() {
		return ErrDeviceAlreadyEnabled
	}
	d.Status = DeviceStatusRegistered
	d.UpdatedAt = time.Now()
	return nil
}

// CanExecuteOperation checks if the device can execute an operation.
// In current MVP stage (no Gateway), this is not used but reserved for future.
func (d *Device) CanExecuteOperation() bool {
	return d.Status.IsActive() && !d.Status.IsDisabled()
}

// GetCapabilities returns the device's capabilities from its DeviceType.
// This method requires the DeviceType to be passed in.
// Device does NOT store capabilities separately - always inherits from DeviceType.
func (d *Device) GetCapabilities(dt *DeviceType) []Capability {
	if dt == nil {
		return []Capability{}
	}
	return dt.Capabilities
}

// HasCapability checks if the device has a specific capability via its DeviceType.
func (d *Device) HasCapability(dt *DeviceType, cap Capability) bool {
	if dt == nil {
		return false
	}
	return dt.HasCapability(cap)
}
