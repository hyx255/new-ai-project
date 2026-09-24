package model

import "testing"

func TestDeviceType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dt      *DeviceType
		wantErr error
	}{
		{
			name: "valid device type",
			dt: &DeviceType{
				ID:           "dt_test1",
				Name:         "IP功放",
				Vendor:       "DSPPA",
				Model:        "MP2806",
				Capabilities: []Capability{CapabilityQueryStatus, CapabilitySetVolume},
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			dt: &DeviceType{
				ID:     "dt_test2",
				Name:   "",
				Vendor: "DSPPA",
				Model:  "MP2806",
			},
			wantErr: ErrDeviceTypeNameEmpty,
		},
		{
			name: "empty vendor",
			dt: &DeviceType{
				ID:     "dt_test3",
				Name:   "IP功放",
				Vendor: "",
				Model:  "MP2806",
			},
			wantErr: ErrDeviceTypeVendorEmpty,
		},
		{
			name: "empty model",
			dt: &DeviceType{
				ID:     "dt_test4",
				Name:   "IP功放",
				Vendor: "DSPPA",
				Model:  "",
			},
			wantErr: ErrDeviceTypeModelEmpty,
		},
		{
			name: "invalid capability",
			dt: &DeviceType{
				ID:           "dt_test5",
				Name:         "IP功放",
				Vendor:       "DSPPA",
				Model:        "MP2806",
				Capabilities: []Capability{Capability("INVALID")},
			},
			wantErr: ErrInvalidCapability,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dt.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() = nil, want %v", tt.wantErr)
				}
			}
		})
	}
}

func TestDeviceType_HasCapability(t *testing.T) {
	dt := &DeviceType{
		Capabilities: []Capability{CapabilityQueryStatus, CapabilitySetVolume},
	}

	if !dt.HasCapability(CapabilityQueryStatus) {
		t.Error("HasCapability(QueryStatus) = false, want true")
	}
	if !dt.HasCapability(CapabilitySetVolume) {
		t.Error("HasCapability(SetVolume) = false, want true")
	}
	if dt.HasCapability(CapabilityRestart) {
		t.Error("HasCapability(Restart) = true, want false")
	}
}

func TestDeviceType_CapabilitiesJSON(t *testing.T) {
	dt := &DeviceType{
		Capabilities: []Capability{CapabilityQueryStatus, CapabilityRestart},
	}

	data, err := dt.CapabilitiesJSON()
	if err != nil {
		t.Fatalf("CapabilitiesJSON() error = %v", err)
	}

	dt2 := &DeviceType{}
	if err := dt2.SetCapabilitiesFromJSON(data); err != nil {
		t.Fatalf("SetCapabilitiesFromJSON() error = %v", err)
	}

	if len(dt2.Capabilities) != 2 {
		t.Errorf("got %d capabilities, want 2", len(dt2.Capabilities))
	}
}
