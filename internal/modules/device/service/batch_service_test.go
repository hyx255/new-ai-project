package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"broadcast-platform/internal/modules/device/adapter"
	"broadcast-platform/internal/modules/device/engine"
	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
)

func setupBatchTestDB(t *testing.T) (*database.DB, *engine.ExecutionEngine, *logging.Logger) {
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

	for _, name := range []string{
		"001_create_device_tables.sql",
		"002_create_operation_tables.sql",
		"003_batch_operation_support.sql",
	} {
		sql, err := os.ReadFile("../../../../migrations/" + name)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", name, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			t.Fatalf("failed to run migration %s: %v", name, err)
		}
	}

	// Create engine for tests
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)
	gw := adapter.NewSimulatorAdapter(logger.WithModule("simulator"))

	eng := engine.NewExecutionEngine(5, opRepo, execRepo, devRepo, gw, logger.WithModule("engine"))

	return db, eng, logger
}

func createTestDevice(t *testing.T, ctx context.Context, dtRepo repository.DeviceTypeRepository, devRepo repository.DeviceRepository, name string, caps []model.Capability, status model.DeviceStatus, address string) *model.Device {
	t.Helper()

	dt := &model.DeviceType{
		ID:           model.GenerateDeviceTypeID(),
		Name:         "DT-" + name,
		Vendor:       "TestVendor",
		Model:        "TestModel",
		Description:  "Test",
		Capabilities: caps,
	}
	if err := dtRepo.Create(ctx, dt); err != nil {
		t.Fatalf("create device type: %v", err)
	}

	dev := &model.Device{
		ID:           model.GenerateDeviceID(),
		Name:         name,
		DeviceTypeID: dt.ID,
		Address:      address,
		Status:       status,
	}
	if err := devRepo.Create(ctx, dev); err != nil {
		t.Fatalf("create device: %v", err)
	}

	return dev
}

// --- Validation Tests ---

func TestBatchService_EmptyDeviceIDs(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(5 * time.Second)

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	_, err := svc.CreateBatchOperation(context.Background(), CreateBatchParams{
		DeviceIDs: []string{},
		Type:      model.OperationTypeQueryStatus,
	})

	if err == nil {
		t.Fatal("expected error for empty device_ids")
	}
	if err != ErrEmptyDeviceIDs {
		t.Fatalf("expected ErrEmptyDeviceIDs, got: %v", err)
	}
}

func TestBatchService_TooManyDevices(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(5 * time.Second)

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 3, logger)

	_, err := svc.CreateBatchOperation(context.Background(), CreateBatchParams{
		DeviceIDs: []string{"d1", "d2", "d3", "d4"},
		Type:      model.OperationTypeQueryStatus,
	})

	if err == nil {
		t.Fatal("expected error for too many devices")
	}
}

func TestBatchService_DuplicateDeviceIDs(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(5 * time.Second)

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	_, err := svc.CreateBatchOperation(context.Background(), CreateBatchParams{
		DeviceIDs: []string{"dev1", "dev2", "dev1"},
		Type:      model.OperationTypeQueryStatus,
	})

	if err == nil {
		t.Fatal("expected error for duplicate device_ids")
	}
}

// --- Pre-check Tests ---

func TestBatchService_AllSkipped_DeviceNotFound(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(5 * time.Second)

	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	result, err := svc.CreateBatchOperation(context.Background(), CreateBatchParams{
		DeviceIDs: []string{"dev_nonexistent_1", "dev_nonexistent_2"},
		Type:      model.OperationTypeQueryStatus,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Operation.Status != model.OperationStatusFailed {
		t.Errorf("expected FAILED, got %s", result.Operation.Status)
	}
	if result.Operation.SkippedCount != 2 {
		t.Errorf("expected 2 skipped, got %d", result.Operation.SkippedCount)
	}
	if result.Operation.TotalCount != 2 {
		t.Errorf("expected total 2, got %d", result.Operation.TotalCount)
	}

	// Check executions are SKIPPED
	for _, exec := range result.Executions {
		if exec.Status != model.ExecutionStatusSkipped {
			t.Errorf("execution %s: expected SKIPPED, got %s", exec.ID, exec.Status)
		}
	}
}

func TestBatchService_AllSkipped_Disabled(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(5 * time.Second)

	ctx := context.Background()
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	dev := createTestDevice(t, ctx, dtRepo, devRepo, "DisabledDev",
		[]model.Capability{model.CapabilityQueryStatus}, model.DeviceStatusDisabled, "192.168.1.1")

	result, err := svc.CreateBatchOperation(ctx, CreateBatchParams{
		DeviceIDs: []string{dev.ID},
		Type:      model.OperationTypeQueryStatus,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Operation.Status != model.OperationStatusFailed {
		t.Errorf("expected FAILED, got %s", result.Operation.Status)
	}
	if result.Operation.SkippedCount != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Operation.SkippedCount)
	}
}

// --- Successful Batch Test ---

func TestBatchService_QueryStatus_AllSuccess(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(10 * time.Second)

	ctx := context.Background()
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	// Create 3 devices
	var deviceIDs []string
	for i := 0; i < 3; i++ {
		dev := createTestDevice(t, ctx, dtRepo, devRepo, "Speaker"+string(rune('A'+i)),
			[]model.Capability{model.CapabilityQueryStatus}, model.DeviceStatusRegistered,
			"192.168.1."+string(rune('0'+i+1)))
		deviceIDs = append(deviceIDs, dev.ID)
	}

	result, err := svc.CreateBatchOperation(ctx, CreateBatchParams{
		DeviceIDs: deviceIDs,
		Type:      model.OperationTypeQueryStatus,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Operation.TotalCount != 3 {
		t.Errorf("expected total 3, got %d", result.Operation.TotalCount)
	}

	// Wait for engine to process
	time.Sleep(3 * time.Second)

	// Check final status
	op, execs, err := svc.GetBatchOperation(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("get batch operation: %v", err)
	}

	if op.Status != model.OperationStatusCompleted {
		t.Errorf("expected COMPLETED, got %s", op.Status)
	}
	if op.SuccessCount != 3 {
		t.Errorf("expected 3 success, got %d", op.SuccessCount)
	}

	for _, exec := range execs {
		if exec.Status != model.ExecutionStatusSuccess {
			t.Errorf("execution %s: expected SUCCESS, got %s (error: %s)", exec.ID, exec.Status, exec.Error)
		}
	}
}

// --- Mixed Results Test ---

func TestBatchService_PartialSuccess(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(10 * time.Second)

	ctx := context.Background()
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	// Create 1 active device and 1 disabled device
	dev1 := createTestDevice(t, ctx, dtRepo, devRepo, "ActiveDev",
		[]model.Capability{model.CapabilityQueryStatus}, model.DeviceStatusRegistered, "192.168.1.10")
	dev2 := createTestDevice(t, ctx, dtRepo, devRepo, "DisabledDev",
		[]model.Capability{model.CapabilityQueryStatus}, model.DeviceStatusDisabled, "192.168.1.11")

	result, err := svc.CreateBatchOperation(ctx, CreateBatchParams{
		DeviceIDs: []string{dev1.ID, dev2.ID},
		Type:      model.OperationTypeQueryStatus,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for engine
	time.Sleep(3 * time.Second)

	op, _, err := svc.GetBatchOperation(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("get batch operation: %v", err)
	}

	// Should be PARTIAL_SUCCESS: 1 success + 1 skipped
	if op.Status != model.OperationStatusPartialSuccess {
		t.Errorf("expected PARTIAL_SUCCESS, got %s", op.Status)
	}
	if op.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", op.SuccessCount)
	}
	if op.SkippedCount != 1 {
		t.Errorf("expected 1 skipped, got %d", op.SkippedCount)
	}
}

// --- Progress Test ---

func TestBatchService_GetProgress(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(10 * time.Second)

	ctx := context.Background()
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	dev := createTestDevice(t, ctx, dtRepo, devRepo, "ProgressDev",
		[]model.Capability{model.CapabilityQueryStatus}, model.DeviceStatusRegistered, "192.168.1.20")

	result, err := svc.CreateBatchOperation(ctx, CreateBatchParams{
		DeviceIDs: []string{dev.ID},
		Type:      model.OperationTypeQueryStatus,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get progress immediately
	progress, err := svc.GetProgress(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}

	if progress.TotalCount != 1 {
		t.Errorf("expected total 1, got %d", progress.TotalCount)
	}

	// Wait for completion
	time.Sleep(3 * time.Second)

	progress, err = svc.GetProgress(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}

	if !progress.IsComplete {
		t.Error("expected IsComplete to be true")
	}
	if progress.Status != model.OperationStatusCompleted {
		t.Errorf("expected COMPLETED, got %s", progress.Status)
	}
}

// --- SET_VOLUME Batch Test ---

func TestBatchService_SetVolume(t *testing.T) {
	db, eng, logger := setupBatchTestDB(t)
	defer db.Close()
	defer eng.Shutdown(10 * time.Second)

	ctx := context.Background()
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	opRepo := repository.NewOperationRepository(db)
	execRepo := repository.NewExecutionRepository(db)

	svc := NewBatchOperationService(devRepo, dtRepo, opRepo, execRepo, eng, 100, logger)

	dev := createTestDevice(t, ctx, dtRepo, devRepo, "VolumeDev",
		[]model.Capability{model.CapabilitySetVolume}, model.DeviceStatusRegistered, "192.168.1.30")

	params, _ := json.Marshal(map[string]int{"volume": 75})

	result, err := svc.CreateBatchOperation(ctx, CreateBatchParams{
		DeviceIDs: []string{dev.ID},
		Type:      model.OperationTypeSetVolume,
		Parameters: params,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for engine
	time.Sleep(3 * time.Second)

	op, execs, err := svc.GetBatchOperation(ctx, result.Operation.ID)
	if err != nil {
		t.Fatalf("get batch operation: %v", err)
	}

	if op.Status != model.OperationStatusCompleted {
		t.Errorf("expected COMPLETED, got %s", op.Status)
	}

	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}
	if execs[0].Status != model.ExecutionStatusSuccess {
		t.Errorf("expected SUCCESS, got %s", execs[0].Status)
	}
}