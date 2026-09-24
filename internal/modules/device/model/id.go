package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// GenerateDeviceTypeID generates a new DeviceType ID in format: dt_<ULID-like>
func GenerateDeviceTypeID() string {
	return fmt.Sprintf("dt_%s", generateULID())
}

// GenerateDeviceID generates a new Device ID in format: dev_<ULID-like>
func GenerateDeviceID() string {
	return fmt.Sprintf("dev_%s", generateULID())
}

// GenerateOperationID generates a new Operation ID in format: op_<ULID-like>
func GenerateOperationID() string {
	return fmt.Sprintf("op_%s", generateULID())
}

// GenerateExecutionID generates a new Execution ID in format: ex_<ULID-like>
func GenerateExecutionID() string {
	return fmt.Sprintf("ex_%s", generateULID())
}

// generateULID generates a ULID-like string (time-based + random).
// Format: 10 chars timestamp (base36) + 16 chars random (base36)
func generateULID() string {
	// Time component: milliseconds since epoch
	now := time.Now().UnixMilli()
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(now))

	// Random component
	randBytes := make([]byte, 10)
	if _, err := rand.Read(randBytes); err != nil {
		// Fallback: use timestamp only (should never happen)
		return fmt.Sprintf("%x%x", timeBytes, time.Now().UnixNano())
	}

	// Encode as hex for simplicity and readability
	return fmt.Sprintf("%x%x", timeBytes[:5], randBytes)
}