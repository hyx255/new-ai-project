package model

import "testing"

func TestCapability_IsValid(t *testing.T) {
	tests := []struct {
		cap  Capability
		want bool
	}{
		{CapabilityQueryStatus, true},
		{CapabilitySetVolume, true},
		{CapabilityRestart, true},
		{Capability("UNKNOWN"), false},
		{Capability(""), false},
		{Capability("query_status"), false}, // case-sensitive
	}

	for _, tt := range tests {
		if got := tt.cap.IsValid(); got != tt.want {
			t.Errorf("Capability(%q).IsValid() = %v, want %v", tt.cap, got, tt.want)
		}
	}
}

func TestAllCapabilities(t *testing.T) {
	caps := AllCapabilities()
	if len(caps) != 3 {
		t.Errorf("AllCapabilities() returned %d capabilities, want 3", len(caps))
	}
	for _, cap := range caps {
		if !cap.IsValid() {
			t.Errorf("AllCapabilities() contains invalid capability: %s", cap)
		}
	}
}
