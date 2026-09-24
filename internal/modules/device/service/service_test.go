package service

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"broadcast-platform/internal/modules/device/model"
	"broadcast-platform/internal/modules/device/repository"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
)

// setupTestDB creates a temporary SQLite database and runs migrations.
func setupTestDB(t *testing.T) *database.DB {
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
		t.Fatalf("failed to create database: %v", err)
	}

	// Run migrations
	migrationSQL, err := os.ReadFile("../../../../migrations/001_create_device_tables.sql")
	if err != nil {
		t.Fatalf("failed to read migration: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}

	return db
}

// ============================================================
// DeviceType CRUD Tests
// ============================================================

func TestDeviceTypeService_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	// Test successful creation
	dt, err := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Description:  "测试设备",
		Capabilities: []model.Capability{model.CapabilityQueryStatus, model.CapabilitySetVolume},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if dt.ID == "" {
		t.Error("Create() returned empty ID")
	}
	if dt.Name != "IP功放" {
		t.Errorf("Name = %v, want IP功放", dt.Name)
	}

	// Test duplicate creation (BR-001)
	_, err = svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err == nil {
		t.Error("Create() should fail for duplicate name+vendor+model")
	}
	if err != ErrDeviceTypeAlreadyExists {
		t.Errorf("Create() error = %v, want ErrDeviceTypeAlreadyExists", err)
	}
}

func TestDeviceTypeService_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	_, err := svc.GetByID(ctx, "dt_nonexistent")
	if err != ErrDeviceTypeNotFound {
		t.Errorf("GetByID() error = %v, want ErrDeviceTypeNotFound", err)
	}
}

// ============================================================
// DeviceType Update Regression Tests
// ============================================================

func TestDeviceTypeService_Update_Name(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	updated, err := svc.Update(ctx, dt.ID, UpdateDeviceTypeParams{
		Name:         "IP功放V2",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err != nil {
		t.Fatalf("Update name error = %v", err)
	}
	if updated.Name != "IP功放V2" {
		t.Errorf("Name = %v, want IP功放V2", updated.Name)
	}
}

func TestDeviceTypeService_Update_Vendor(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	updated, err := svc.Update(ctx, dt.ID, UpdateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA-V2",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err != nil {
		t.Fatalf("Update vendor error = %v", err)
	}
	if updated.Vendor != "DSPPA-V2" {
		t.Errorf("Vendor = %v, want DSPPA-V2", updated.Vendor)
	}
}

func TestDeviceTypeService_Update_Model(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	updated, err := svc.Update(ctx, dt.ID, UpdateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2807",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err != nil {
		t.Fatalf("Update model error = %v", err)
	}
	if updated.Model != "MP2807" {
		t.Errorf("Model = %v, want MP2807", updated.Model)
	}
}

func TestDeviceTypeService_Update_Description(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	updated, err := svc.Update(ctx, dt.ID, UpdateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Description:  "updated description",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err != nil {
		t.Fatalf("Update description error = %v", err)
	}
	if updated.Description != "updated description" {
		t.Errorf("Description = %v, want 'updated description'", updated.Description)
	}
}

func TestDeviceTypeService_Update_Capabilities(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	updated, err := svc.Update(ctx, dt.ID, UpdateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus, model.CapabilitySetVolume, model.CapabilityRestart},
	})
	if err != nil {
		t.Fatalf("Update capabilities error = %v", err)
	}
	if len(updated.Capabilities) != 3 {
		t.Errorf("Capabilities count = %d, want 3", len(updated.Capabilities))
	}
}

func TestDeviceTypeService_Update_UniqueConflict(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	// Create two device types
	svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放A",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	dt2, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放B",
		Vendor:       "DSPPA",
		Model:        "MP2807",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	// Try to update dt2 to conflict with dt1
	_, err := svc.Update(ctx, dt2.ID, UpdateDeviceTypeParams{
		Name:         "IP功放A",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err == nil {
		t.Error("Update() should fail for unique conflict")
	}
	if err != ErrDeviceTypeAlreadyExists {
		t.Errorf("Update() error = %v, want ErrDeviceTypeAlreadyExists", err)
	}
}

func TestDeviceTypeService_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	_, err := svc.Update(ctx, "dt_nonexistent", UpdateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	if err == nil {
		t.Error("Update() should fail for non-existent ID")
	}
	if err != ErrDeviceTypeNotFound {
		t.Errorf("Update() error = %v, want ErrDeviceTypeNotFound", err)
	}
}

// ============================================================
// DeviceType Delete Tests
// ============================================================

func TestDeviceTypeService_Delete_WithDevices(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	dtSvc := NewDeviceTypeService(dtRepo, logger)

	ctx := context.Background()

	dt, _ := dtSvc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	dev, _ := model.NewDevice("dev_test1", "测试设备", dt.ID, "localhost:9001")
	if err := devRepo.Create(ctx, dev); err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	err := dtSvc.Delete(ctx, dt.ID)
	if err == nil {
		t.Error("Delete() should fail when devices exist")
	}
	if err != ErrDeviceTypeInUse {
		t.Errorf("Delete() error = %v, want ErrDeviceTypeInUse", err)
	}
}

func TestDeviceTypeService_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	err := svc.Delete(ctx, "dt_nonexistent")
	if err != ErrDeviceTypeNotFound {
		t.Errorf("Delete() error = %v, want ErrDeviceTypeNotFound", err)
	}
}

func TestDeviceTypeService_Delete_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	repo := repository.NewDeviceTypeRepository(db)
	svc := NewDeviceTypeService(repo, logger)

	ctx := context.Background()

	dt, _ := svc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	err := svc.Delete(ctx, dt.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = svc.GetByID(ctx, dt.ID)
	if err != ErrDeviceTypeNotFound {
		t.Errorf("GetByID after delete error = %v, want ErrDeviceTypeNotFound", err)
	}
}

// ============================================================
// Device Tests
// ============================================================

func TestDeviceService_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	dtSvc := NewDeviceTypeService(dtRepo, logger)
	devSvc := NewDeviceService(devRepo, dtRepo, logger)

	ctx := context.Background()

	dt, _ := dtSvc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})

	dev, err := devSvc.Create(ctx, CreateDeviceParams{
		Name:         "测试设备",
		DeviceTypeID: dt.ID,
		Address:      "simulator://localhost:9001",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if dev.Status != model.DeviceStatusRegistered {
		t.Errorf("Status = %v, want REGISTERED", dev.Status)
	}

	// Create device with non-existent type
	_, err = devSvc.Create(ctx, CreateDeviceParams{
		Name:         "测试设备2",
		DeviceTypeID: "dt_nonexistent",
	})
	if err == nil {
		t.Error("Create() should fail for non-existent device type")
	}
	if err != ErrDeviceTypeInvalid {
		t.Errorf("Create() error = %v, want ErrDeviceTypeInvalid", err)
	}
}

func TestDeviceService_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	devSvc := NewDeviceService(devRepo, dtRepo, logger)

	ctx := context.Background()

	_, err := devSvc.GetByID(ctx, "dev_nonexistent")
	if err != ErrDeviceNotFound {
		t.Errorf("GetByID() error = %v, want ErrDeviceNotFound", err)
	}
}

func TestDeviceService_StatusTransitions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "json"})
	dtRepo := repository.NewDeviceTypeRepository(db)
	devRepo := repository.NewDeviceRepository(db)
	dtSvc := NewDeviceTypeService(dtRepo, logger)
	devSvc := NewDeviceService(devRepo, dtRepo, logger)

	ctx := context.Background()

	dt, _ := dtSvc.Create(ctx, CreateDeviceTypeParams{
		Name:         "IP功放",
		Vendor:       "DSPPA",
		Model:        "MP2806",
		Capabilities: []model.Capability{model.CapabilityQueryStatus},
	})
	dev, _ := devSvc.Create(ctx, CreateDeviceParams{
		Name:         "测试设备",
		DeviceTypeID: dt.ID,
		Address:      "localhost:9001",
	})

	// REGISTERED -> DISABLED
	dev, err := devSvc.Disable(ctx, dev.ID)
	if err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	if dev.Status != model.DeviceStatusDisabled {
		t.Errorf("Status = %v, want DISABLED", dev.Status)
	}

	// DISABLED -> REGISTERED (enable)
	dev, err = devSvc.Enable(ctx, dev.ID)
	if err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if dev.Status != model.DeviceStatusRegistered {
		t.Errorf("Status = %v, want REGISTERED", dev.Status)
	}

	// Try to enable when already enabled
	_, err = devSvc.Enable(ctx, dev.ID)
	if err == nil {
		t.Error("Enable() should fail when device is not DISABLED")
	}

	// Disable again
	_, err = devSvc.Disable(ctx, dev.ID)
	if err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	// Try to disable when already disabled
	_, err = devSvc.Disable(ctx, dev.ID)
	if err == nil {
		t.Error("Disable() should fail when device is already DISABLED")
	}
}

// Ensure sql import is used
var _ = sql.ErrNoRows
