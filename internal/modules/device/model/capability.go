package model

// Capability represents a device capability that can be executed.
// Capabilities are defined by DeviceType and inherited by Device.
type Capability string

const (
	CapabilityQueryStatus Capability = "QUERY_STATUS"
	CapabilitySetVolume   Capability = "SET_VOLUME"
	CapabilityRestart     Capability = "RESTART"
)

// IsValid checks if the capability is a known MVP capability.
func (c Capability) IsValid() bool {
	switch c {
	case CapabilityQueryStatus, CapabilitySetVolume, CapabilityRestart:
		return true
	}
	return false
}

// AllCapabilities returns all valid MVP capabilities.
func AllCapabilities() []Capability {
	return []Capability{
		CapabilityQueryStatus,
		CapabilitySetVolume,
		CapabilityRestart,
	}
}
