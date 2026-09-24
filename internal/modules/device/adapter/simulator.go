package adapter

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"broadcast-platform/internal/modules/device/gateway"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

// SimulatorState represents the internal state of a simulated device.
type SimulatorState struct {
	Online     bool
	Volume     int
	Uptime     int
	Restarting bool
}

// SimulatorAdapter implements DeviceGateway for testing purposes.
// It maintains in-memory state for each device address.
type SimulatorAdapter struct {
	mu     sync.RWMutex
	states map[string]*SimulatorState
	logger *logging.Logger
}

// NewSimulatorAdapter creates a new simulator adapter.
func NewSimulatorAdapter(logger *logging.Logger) *SimulatorAdapter {
	return &SimulatorAdapter{
		states: make(map[string]*SimulatorState),
		logger: logger,
	}
}

// getOrCreateState gets or creates simulator state for a device.
func (s *SimulatorAdapter) getOrCreateState(device *model.Device) *SimulatorState {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := device.Address
	if state, exists := s.states[addr]; exists {
		return state
	}

	// Create default state
	state := &SimulatorState{
		Online: true,
		Volume: 50,
		Uptime: 3600,
	}
	s.states[addr] = state
	return state
}

// simulateLatency adds random latency to simulate network delay.
func simulateLatency(minMs, maxMs int) {
	latency := time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond
	time.Sleep(latency)
}

// QueryStatus queries the simulated device status.
func (s *SimulatorAdapter) QueryStatus(ctx context.Context, device *model.Device) (*gateway.QueryStatusResult, error) {
	s.logger.Debug("simulator: query status",
		zap.String("device_id", device.ID),
		zap.String("address", device.Address),
	)

	// Special address for timeout simulation
	if device.Address == "simulator://timeout" {
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("%w: query timeout", gateway.ErrDeviceTimeout)
	}

	// Special address for error simulation
	if device.Address == "simulator://error" {
		return nil, fmt.Errorf("%w: simulated device error", gateway.ErrDeviceError)
	}

	// Special address for offline simulation
	if device.Address == "simulator://offline" {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	// Simulate network latency (50-200ms)
	simulateLatency(50, 200)

	state := s.getOrCreateState(device)
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !state.Online {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	return &gateway.QueryStatusResult{
		Online: state.Online,
		Volume: state.Volume,
		Uptime: state.Uptime,
	}, nil
}

// SetVolume sets the simulated device volume.
func (s *SimulatorAdapter) SetVolume(ctx context.Context, device *model.Device, volume int) (*gateway.SetVolumeResult, error) {
	s.logger.Debug("simulator: set volume",
		zap.String("device_id", device.ID),
		zap.String("address", device.Address),
		zap.Int("volume", volume),
	)

	// Special address for timeout simulation
	if device.Address == "simulator://timeout" {
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("%w: set volume timeout", gateway.ErrDeviceTimeout)
	}

	// Special address for error simulation
	if device.Address == "simulator://error" {
		return nil, fmt.Errorf("%w: simulated device error", gateway.ErrDeviceError)
	}

	// Special address for offline simulation
	if device.Address == "simulator://offline" {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	// Simulate network latency (100-300ms)
	simulateLatency(100, 300)

	state := s.getOrCreateState(device)
	s.mu.Lock()
	defer s.mu.Unlock()

	if !state.Online {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	previousVolume := state.Volume
	state.Volume = volume

	return &gateway.SetVolumeResult{
		PreviousVolume: previousVolume,
		CurrentVolume:  volume,
	}, nil
}

// Restart restarts the simulated device.
func (s *SimulatorAdapter) Restart(ctx context.Context, device *model.Device) (*gateway.RestartResult, error) {
	s.logger.Debug("simulator: restart",
		zap.String("device_id", device.ID),
		zap.String("address", device.Address),
	)

	// Special address for timeout simulation
	if device.Address == "simulator://timeout" {
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("%w: restart timeout", gateway.ErrDeviceTimeout)
	}

	// Special address for error simulation
	if device.Address == "simulator://error" {
		return nil, fmt.Errorf("%w: simulated device error", gateway.ErrDeviceError)
	}

	// Special address for offline simulation
	if device.Address == "simulator://offline" {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	// Simulate network latency (500-1000ms)
	simulateLatency(500, 1000)

	state := s.getOrCreateState(device)
	s.mu.Lock()
	defer s.mu.Unlock()

	if !state.Online {
		return nil, fmt.Errorf("%w: device is offline", gateway.ErrDeviceOffline)
	}

	// Simulate restart: go offline, then back online after 2 seconds
	state.Online = false
	state.Restarting = true

	go func() {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		defer s.mu.Unlock()
		state.Online = true
		state.Restarting = false
		state.Uptime = 0
	}()

	return &gateway.RestartResult{
		Success: true,
		Message: "restart initiated",
	}, nil
}