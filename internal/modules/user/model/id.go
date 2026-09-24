package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// GenerateUserID generates a new User ID in format: usr_<hex>
func GenerateUserID() string {
	return fmt.Sprintf("usr_%s", generateULID())
}

func generateULID() string {
	now := time.Now().UnixMilli()
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(now))

	randBytes := make([]byte, 10)
	if _, err := rand.Read(randBytes); err != nil {
		return fmt.Sprintf("%x%x", timeBytes, time.Now().UnixNano())
	}

	return fmt.Sprintf("%x%x", timeBytes[:5], randBytes)
}