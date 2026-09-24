package gateway

import (
	"context"
	"errors"

	"broadcast-platform/internal/modules/device/model"
)

// Standard gateway error codes
var (
	ErrDeviceOffline    = errors.New("device offline")
	ErrDeviceTimeout    = errors.New("device timeout")
	ErrDeviceError      = errors.New("device error")
	ErrUnsupported      = errors.New("capability not supported")
)

// QueryStatusResult represents the result of a status query.
type QueryStatusResult struct {
	Online bool `json:"online"`
	Volume int  `json:"volume"`
	Uptime int  `json:"uptime"`
}

// SetVolumeResult represents the result of a volume set operation.
type SetVolumeResult struct {
	PreviousVolume int `json:"previous_volume"`
	CurrentVolume  int `json:"current_volume"`
}

// RestartResult represents the result of a restart operation.
type RestartResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeviceGateway defines the abstract interface for device capabilities.
// Service layer depends on this interface, not on concrete adapter implementations.
type DeviceGateway interface {
	// QueryStatus queries the current status of a device.
	QueryStatus(ctx context.Context, device *model.Device) (*QueryStatusResult, error)

	// SetVolume sets the device volume (0-100).
	SetVolume(ctx context.Context, device *model.Device, volume int) (*SetVolumeResult, error)

	// Restart restarts the device.
	Restart(ctx context.Context, device *model.Device) (*RestartResult, error)
}