package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"broadcast-platform/internal/modules/device/adapter"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
)

// setupOperationTestDB creates a test database with all required tables.
func setupOperationTestDB(t *testing.T) *database.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := config.DatabaseConfig{
		Driver: "sqlite3",
		DSN:    dbPath,
	}

	logger, err := logging.New(config.LoggingConfig{
		Level:  "error",
		Format: "json",
	})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	db, err := database.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Run migrations
	migrationSQL1, err := os.ReadFile("../../../../migrations/001_create_device_tables.sql")
	if err != nil {
		t.Fatalf("failed to read migration 001: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL1)); err != nil {
		t.Fatalf("failed to run migration 001: %v", err)
	}

	migrationSQL2, err := os.ReadFile("../../../../migrations/002_create_operation_tables.sql")
	if err != nil {
		t.Fatalf("failed to read migration 002: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL2)); err != nil {
		t.Fatalf("failed to run migration 002: %v", err)
	}

	migrationSQL3, err := os.ReadFile("../../../../migrations/003_batch_operation_support.sql")
	if err != nil {
		t.Fatalf("failed to read migration 003: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL3)); err != nil {
		t.Fatalf("failed to run migration 003: %v", err)
	}

	return db
}

// createTestDeviceWithCapabilities creates a test device type and device with specified capabilities.
func createTestDeviceWithCapabilities(
	ctx context.Context,
	dtRepo repository.DeviceTypeRepository,
	devRepo repository.DeviceRepository,
	capabilities []model.Capability,
) (*model.Device, error) {
	dt := &model.DeviceType{
		ID:           model.GenerateDeviceTypeID(),
		Name:         "Test Device Type",
		Vendor:       "Test Vendor",
		Model:        "Test Model",
		Description:  "Test Description",
		Capabilities: capabilities,
	}

	if err := dtRepo.Create(ctx, dt); err != nil {
		return nil, err
	}

	dev := &model.Device{
		ID:           model.GenerateDeviceID(),
		Name:         "Test Device",
		DeviceTypeID: dt.ID,
		Address:      "192.168.1.100",
		Status:       model.DeviceStatusRegistered,
	}

	if err := devRepo.Create(ctx, dev); err != nil {
		return nil, err
	}

	return dev, nil
}

func TestOperationService_Create_QueryStatus(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device with QUERY_STATUS capability
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Create operation
	params := CreateOperationParams{
		DeviceID: dev.ID,
		Type:     model.OperationTypeQueryStatus,
	}

	result, err := svc.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify operation completed
	if result.Operation.Status != model.OperationStatusCompleted {
		t.Errorf("Operation.Status = %v, want %v", result.Operation.Status, model.OperationStatusCompleted)
	}

	// Verify execution succeeded
	if result.Execution.Status != model.ExecutionStatusSuccess {
		t.Errorf("Execution.Status = %v, want %v", result.Execution.Status, model.ExecutionStatusSuccess)
	}

	// Verify device status updated to ACTIVE
	updatedDev, err := devRepo.GetByID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("failed to get updated device: %v", err)
	}

	if updatedDev.Status != model.DeviceStatusActive {
		t.Errorf("Device.Status = %v, want %v", updatedDev.Status, model.DeviceStatusActive)
	}
}

func TestOperationService_Create_SetVolume(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device with SET_VOLUME capability
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilitySetVolume})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Create operation with volume parameter
	paramsJSON, _ := json.Marshal(SetVolumeParams{Volume: 75})
	params := CreateOperationParams{
		DeviceID:   dev.ID,
		Type:       model.OperationTypeSetVolume,
		Parameters: paramsJSON,
	}

	result, err := svc.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify operation completed
	if result.Operation.Status != model.OperationStatusCompleted {
		t.Errorf("Operation.Status = %v, want %v", result.Operation.Status, model.OperationStatusCompleted)
	}

	// Verify execution succeeded
	if result.Execution.Status != model.ExecutionStatusSuccess {
		t.Errorf("Execution.Status = %v, want %v", result.Execution.Status, model.ExecutionStatusSuccess)
	}

	// Verify result contains volume
	var setResult map[string]interface{}
	if err := json.Unmarshal(result.Execution.Result, &setResult); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if currentVolume, ok := setResult["current_volume"].(float64); !ok || int(currentVolume) != 75 {
		t.Errorf("Result.current_volume = %v, want 75", setResult["current_volume"])
	}
}

func TestOperationService_Create_SetVolume_InvalidParameter(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device with SET_VOLUME capability
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilitySetVolume})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Try to set volume > 100
	paramsJSON, _ := json.Marshal(SetVolumeParams{Volume: 150})
	params := CreateOperationParams{
		DeviceID:   dev.ID,
		Type:       model.OperationTypeSetVolume,
		Parameters: paramsJSON,
	}

	_, err = svc.Create(ctx, params)
	if err == nil {
		t.Fatal("Create() should fail with invalid volume")
	}

	if !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("Create() error = %v, want ErrInvalidParameter", err)
	}
}

func TestOperationService_Create_DeviceDisabled(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device and disable it
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Disable device
	dev.Status = model.DeviceStatusDisabled
	if err := devRepo.Update(ctx, dev); err != nil {
		t.Fatalf("failed to disable device: %v", err)
	}

	// Try to create operation
	params := CreateOperationParams{
		DeviceID: dev.ID,
		Type:     model.OperationTypeQueryStatus,
	}

	_, err = svc.Create(ctx, params)
	if err == nil {
		t.Fatal("Create() should fail with disabled device")
	}

	if !errors.Is(err, ErrDeviceDisabled) {
		t.Errorf("Create() error = %v, want ErrDeviceDisabled", err)
	}
}

func TestOperationService_Create_CapabilityNotSupported(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device without SET_VOLUME capability
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Try to set volume
	paramsJSON, _ := json.Marshal(SetVolumeParams{Volume: 50})
	params := CreateOperationParams{
		DeviceID:   dev.ID,
		Type:       model.OperationTypeSetVolume,
		Parameters: paramsJSON,
	}

	_, err = svc.Create(ctx, params)
	if err == nil {
		t.Fatal("Create() should fail when capability not supported")
	}

	if !errors.Is(err, ErrCapabilityRequired) {
		t.Errorf("Create() error = %v, want ErrCapabilityRequired", err)
	}
}

func TestOperationService_Create_QueryStatus_DeviceOffline(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device with special offline address
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	dev.Address = "simulator://offline"
	if err := devRepo.Update(ctx, dev); err != nil {
		t.Fatalf("failed to update device address: %v", err)
	}

	// Create operation
	params := CreateOperationParams{
		DeviceID: dev.ID,
		Type:     model.OperationTypeQueryStatus,
	}

	result, err := svc.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Operation should be FAILED when execution fails
	if result.Operation.Status != model.OperationStatusFailed {
		t.Errorf("Operation.Status = %v, want %v", result.Operation.Status, model.OperationStatusFailed)
	}

	// Execution should fail
	if result.Execution.Status != model.ExecutionStatusFailed {
		t.Errorf("Execution.Status = %v, want %v", result.Execution.Status, model.ExecutionStatusFailed)
	}

	// Device status should be updated to OFFLINE
	updatedDev, err := devRepo.GetByID(ctx, dev.ID)
	if err != nil {
		t.Fatalf("failed to get updated device: %v", err)
	}

	if updatedDev.Status != model.DeviceStatusOffline {
		t.Errorf("Device.Status = %v, want %v", updatedDev.Status, model.DeviceStatusOffline)
	}
}

func TestOperationService_GetOperationWithExecutions(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device and operation
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	params := CreateOperationParams{
		DeviceID: dev.ID,
		Type:     model.OperationTypeQueryStatus,
	}

	result, err := svc.Create(ctx, params)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get operation with executions
	op, execs, err := svc.GetOperationWithExecutions(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("GetOperationWithExecutions() error = %v", err)
	}

	if op.ID != result.Operation.ID {
		t.Errorf("Operation.ID = %v, want %v", op.ID, result.Operation.ID)
	}

	if len(execs) != 1 {
		t.Errorf("len(executions) = %v, want 1", len(execs))
	}

	if execs[0].ID != result.Execution.ID {
		t.Errorf("Execution.ID = %v, want %v", execs[0].ID, result.Execution.ID)
	}
}

func TestOperationService_ListDeviceOperations(t *testing.T) {
	db := setupOperationTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	ctx := context.Background()

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger)

	svc := NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger)

	// Create device
	dev, err := createTestDeviceWithCapabilities(ctx, dtRepo, devRepo, []model.Capability{model.CapabilityQueryStatus})
	if err != nil {
		t.Fatalf("failed to create test device: %v", err)
	}

	// Create 3 operations
	for i := 0; i < 3; i++ {
		params := CreateOperationParams{
			DeviceID: dev.ID,
			Type:     model.OperationTypeQueryStatus,
		}

		_, err := svc.Create(ctx, params)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// List operations
	ops, total, err := svc.ListDeviceOperations(ctx, dev.ID, 1, 10)
	if err != nil {
		t.Fatalf("ListDeviceOperations() error = %v", err)
	}

	if total != 3 {
		t.Errorf("total = %v, want 3", total)
	}

	if len(ops) != 3 {
		t.Errorf("len(operations) = %v, want 3", len(ops))
	}
}