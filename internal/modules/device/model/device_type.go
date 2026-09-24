package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrDeviceTypeNameEmpty   = errors.New("device type name is required")
	ErrDeviceTypeVendorEmpty = errors.New("device type vendor is required")
	ErrDeviceTypeModelEmpty  = errors.New("device type model is required")
	ErrInvalidCapability     = errors.New("invalid capability")
)

// DeviceType represents a category of devices with shared capabilities.
type DeviceType struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Vendor       string       `json:"vendor"`
	Model        string       `json:"model"`
	Description  string       `json:"description"`
	Capabilities []Capability `json:"capabilities"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// NewDeviceType creates a new DeviceType with validation.
func NewDeviceType(id, name, vendor, model, description string, capabilities []Capability) (*DeviceType, error) {
	dt := &DeviceType{
		ID:           id,
		Name:         name,
		Vendor:       vendor,
		Model:        model,
		Description:  description,
		Capabilities: capabilities,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := dt.Validate(); err != nil {
		return nil, err
	}

	return dt, nil
}

// Validate checks DeviceType business rules.
func (dt *DeviceType) Validate() error {
	if dt.Name == "" {
		return ErrDeviceTypeNameEmpty
	}
	if dt.Vendor == "" {
		return ErrDeviceTypeVendorEmpty
	}
	if dt.Model == "" {
		return ErrDeviceTypeModelEmpty
	}

	// Validate all capabilities
	for _, cap := range dt.Capabilities {
		if !cap.IsValid() {
			return fmt.Errorf("%w: %s", ErrInvalidCapability, cap)
		}
	}

	return nil
}

// UniqueKey returns the unique identifier for DeviceType (name + vendor + model).
func (dt *DeviceType) UniqueKey() string {
	return fmt.Sprintf("%s|%s|%s", dt.Name, dt.Vendor, dt.Model)
}

// CanRemoveCapability checks if a capability can be safely removed.
// In MVP, we don't track which devices use which capabilities,
// so we allow removal only if no devices reference this DeviceType.
// This check should be done at service layer with device count.
func (dt *DeviceType) CanRemoveCapability(cap Capability) bool {
	// Always allow adding capabilities
	for _, existing := range dt.Capabilities {
		if existing == cap {
			return false // Already exists, not removing
		}
	}
	return true
}

// HasCapability checks if the DeviceType has a specific capability.
func (dt *DeviceType) HasCapability(cap Capability) bool {
	for _, c := range dt.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// CapabilitiesJSON converts capabilities to JSON for database storage.
func (dt *DeviceType) CapabilitiesJSON() ([]byte, error) {
	return json.Marshal(dt.Capabilities)
}

// SetCapabilitiesFromJSON parses capabilities from JSON.
func (dt *DeviceType) SetCapabilitiesFromJSON(data []byte) error {
	if len(data) == 0 {
		dt.Capabilities = []Capability{}
		return nil
	}
	return json.Unmarshal(data, &dt.Capabilities)
}
