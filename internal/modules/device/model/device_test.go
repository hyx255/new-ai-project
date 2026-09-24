package model

import "testing"

func TestDevice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		device  *Device
		wantErr error
	}{
		{
			name: "valid device",
			device: &Device{
				ID:           "dev_test1",
				Name:         "测试设备",
				DeviceTypeID: "dt_test1",
				Status:       DeviceStatusRegistered,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			device: &Device{
				ID:           "dev_test2",
				Name:         "",
				DeviceTypeID: "dt_test1",
			},
			wantErr: ErrDeviceNameEmpty,
		},
		{
			name: "empty device type ID",
			device: &Device{
				ID:           "dev_test3",
				Name:         "测试设备",
				DeviceTypeID: "",
			},
			wantErr: ErrDeviceTypeIDEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()
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

func TestDevice_Disable(t *testing.T) {
	tests := []struct {
		name       string
		status     DeviceStatus
		wantErr    error
		wantStatus DeviceStatus
	}{
		{
			name:       "disable from REGISTERED",
			status:     DeviceStatusRegistered,
			wantErr:    nil,
			wantStatus: DeviceStatusDisabled,
		},
		{
			name:       "disable from ACTIVE",
			status:     DeviceStatusActive,
			wantErr:    nil,
			wantStatus: DeviceStatusDisabled,
		},
		{
			name:       "disable from OFFLINE",
			status:     DeviceStatusOffline,
			wantErr:    nil,
			wantStatus: DeviceStatusDisabled,
		},
		{
			name:       "disable from DISABLED",
			status:     DeviceStatusDisabled,
			wantErr:    ErrDeviceAlreadyDisabled,
			wantStatus: DeviceStatusDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				ID:     "dev_test",
				Status: tt.status,
			}
			err := d.Disable()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Disable() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Disable() = nil, want %v", tt.wantErr)
				}
			}
			if d.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", d.Status, tt.wantStatus)
			}
		})
	}
}

func TestDevice_Enable(t *testing.T) {
	tests := []struct {
		name       string
		status     DeviceStatus
		wantErr    error
		wantStatus DeviceStatus
	}{
		{
			name:       "enable from DISABLED",
			status:     DeviceStatusDisabled,
			wantErr:    nil,
			wantStatus: DeviceStatusRegistered,
		},
		{
			name:       "enable from REGISTERED",
			status:     DeviceStatusRegistered,
			wantErr:    ErrDeviceAlreadyEnabled,
			wantStatus: DeviceStatusRegistered,
		},
		{
			name:       "enable from ACTIVE",
			status:     DeviceStatusActive,
			wantErr:    ErrDeviceAlreadyEnabled,
			wantStatus: DeviceStatusActive,
		},
		{
			name:       "enable from OFFLINE",
			status:     DeviceStatusOffline,
			wantErr:    ErrDeviceAlreadyEnabled,
			wantStatus: DeviceStatusOffline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				ID:     "dev_test",
				Status: tt.status,
			}
			err := d.Enable()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Enable() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Enable() = nil, want %v", tt.wantErr)
				}
			}
			if d.Status != tt.wantStatus {
				t.Errorf("Status = %v, want %v", d.Status, tt.wantStatus)
			}
		})
	}
}

func TestDeviceStatus_IsValid(t *testing.T) {
	tests := []struct {
		status DeviceStatus
		want   bool
	}{
		{DeviceStatusRegistered, true},
		{DeviceStatusActive, true},
		{DeviceStatusOffline, true},
		{DeviceStatusDisabled, true},
		{DeviceStatus("UNKNOWN"), false},
		{DeviceStatus(""), false},
	}

	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("DeviceStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
		}
	}
}
