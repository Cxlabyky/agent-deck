package ledger

import (
	"os"
	"testing"

	"github.com/asheshgoplani/agent-deck/internal/database"
)

func TestGetManager(t *testing.T) {
	// GetManager should return the same instance
	mgr1 := GetManager()
	mgr2 := GetManager()

	if mgr1 != mgr2 {
		t.Error("GetManager should return the same singleton instance")
	}

	if mgr1.baseDir == "" {
		t.Error("baseDir should be set")
	}
}

func TestManagerGetDB(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "ledger-mgr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a manager with custom base dir
	mgr := &Manager{
		databases: make(map[string]*database.DB),
		baseDir:   tmpDir,
	}

	projectPath := "/test/project/path"

	// Get DB - should create it
	db1, err := mgr.GetDB(projectPath)
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	if db1 == nil {
		t.Fatal("db should not be nil")
	}

	// Get DB again - should return cached version
	db2, err := mgr.GetDB(projectPath)
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}
	if db1 != db2 {
		t.Error("should return the same cached DB instance")
	}

	// Close all
	if err := mgr.CloseAll(); err != nil {
		t.Fatalf("failed to close all: %v", err)
	}

	// Map should be empty after close
	if len(mgr.databases) != 0 {
		t.Errorf("databases map should be empty after CloseAll, got %d", len(mgr.databases))
	}
}

func TestManagerIsInitialized(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "ledger-mgr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := &Manager{
		databases: make(map[string]*database.DB),
		baseDir:   tmpDir,
	}

	projectPath := "/test/project/path"

	// Should not be initialized initially
	if mgr.IsInitialized(projectPath) {
		t.Error("project should not be initialized before GetDB")
	}

	// Initialize by getting DB
	_, err = mgr.GetDB(projectPath)
	if err != nil {
		t.Fatalf("failed to get DB: %v", err)
	}

	// Should be initialized now
	if !mgr.IsInitialized(projectPath) {
		t.Error("project should be initialized after GetDB")
	}

	mgr.CloseAll()
}
