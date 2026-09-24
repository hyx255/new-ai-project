package model

import "time"

// OperationType represents the type of device operation.
type OperationType string

const (
	OperationTypeQueryStatus OperationType = "QUERY_STATUS"
	OperationTypeSetVolume   OperationType = "SET_VOLUME"
	OperationTypeRestart     OperationType = "RESTART"
)

// IsValid checks if the operation type is a known MVP type.
func (t OperationType) IsValid() bool {
	switch t {
	case OperationTypeQueryStatus, OperationTypeSetVolume, OperationTypeRestart:
		return true
	}
	return false
}

// ToCapability maps operation type to device capability.
func (t OperationType) ToCapability() Capability {
	switch t {
	case OperationTypeQueryStatus:
		return CapabilityQueryStatus
	case OperationTypeSetVolume:
		return CapabilitySetVolume
	case OperationTypeRestart:
		return CapabilityRestart
	default:
		return ""
	}
}

// IsWriteOperation returns true if this operation modifies device state.
func (t OperationType) IsWriteOperation() bool {
	return t != OperationTypeQueryStatus
}

// MaxRetries returns the maximum retry count for this operation type.
func (t OperationType) MaxRetries() int {
	switch t {
	case OperationTypeQueryStatus:
		return 0
	case OperationTypeSetVolume:
		return 3
	case OperationTypeRestart:
		return 1
	default:
		return 0
	}
}

// RetryInterval returns the fixed retry interval for this operation type.
func (t OperationType) RetryInterval() time.Duration {
	switch t {
	case OperationTypeQueryStatus:
		return 0
	case OperationTypeSetVolume:
		return 1 * time.Second
	case OperationTypeRestart:
		return 2 * time.Second
	default:
		return 1 * time.Second
	}
}