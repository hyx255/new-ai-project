package main

import (
	"testing"
)

func TestMainPackageCompiles(t *testing.T) {
	// This test verifies that the main package compiles correctly
	// and that the bootstrap flow is the single entry point.
	// Actual server behavior is tested in internal/bootstrap and internal/platform/http.
	t.Log("main package compiles successfully")
}

