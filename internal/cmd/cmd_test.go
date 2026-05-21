package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.life-pay.ru/ai/kaiten-cli/internal/db"
)

// ---------------------------------------------------------------------------
//  Sync flag validation
// ---------------------------------------------------------------------------

func TestSyncCmd_NoScope(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"sync"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for sync with no scope, got nil")
	}
	if !strings.Contains(err.Error(), "exactly one of") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSyncCmd_MultipleScopes(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"sync", "--board-id", "1", "--space-id", "2"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for sync with multiple scopes, got nil")
	}
	if !strings.Contains(err.Error(), "exactly one of") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSyncCmd_BoardIDAndAll(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"sync", "--board-id", "1", "--all"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for sync with multiple scopes, got nil")
	}
}

// ---------------------------------------------------------------------------
//  DB status
// ---------------------------------------------------------------------------

func TestDBStatus_NoDatabase(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "nonexistent", "kaiten.db"))

	rootCmd.SetArgs([]string{"db", "status"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("db status should not error on missing db: %v", err)
	}
}

func TestDBStatus_DoesNotRequireAPICredentials(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "")
	t.Setenv("KAITEN_URL", "")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "nonexistent", "kaiten.db"))

	rootCmd.SetArgs([]string{"db", "status"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("db status should not require API credentials: %v", err)
	}
}

func TestDBStatus_WithDatabase(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)

	// Create a real database.
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	database.Close()

	rootCmd.SetArgs([]string{"db", "status"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("db status should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  DB reset
// ---------------------------------------------------------------------------

func TestDBReset_NoYesFlag(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"db", "reset"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for db reset without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected --yes in error, got: %v", err)
	}
}

func TestDBReset_WithYesRemovesDB(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)

	// Create a real database file.
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	database.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("test db should exist after Open")
	}

	rootCmd.SetArgs([]string{"db", "reset", "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("db reset --yes should not error: %v", err)
	}

	// Verify the database file was removed.
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Errorf("database should have been removed after reset, still exists")
	}
}

func TestDBReset_WithYesOnMissingDB(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "no-such-db.kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)

	rootCmd.SetArgs([]string{"db", "reset", "--yes"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("db reset --yes on missing db should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  DB vacuum
// ---------------------------------------------------------------------------

func TestDBVacuum_WithDatabase(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	dbPath := filepath.Join(t.TempDir(), "kaiten.db")
	t.Setenv("KAITEN_DB_PATH", dbPath)

	// Create a real database.
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	database.Close()

	rootCmd.SetArgs([]string{"db", "vacuum"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("db vacuum should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
//  Existing commands still work
// ---------------------------------------------------------------------------

func TestHelpOutput(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	// Just verify that help renders without error.
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("help should not error: %v", err)
	}
}

func TestSyncHelpOutput(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"sync", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("sync --help should not error: %v", err)
	}
}

func TestDBHelpOutput(t *testing.T) {
	t.Setenv("KAITEN_API_TOKEN", "test-token")
	t.Setenv("KAITEN_URL", "https://test.kaiten.ru")
	t.Setenv("KAITEN_DB_PATH", filepath.Join(t.TempDir(), "kaiten.db"))

	rootCmd.SetArgs([]string{"db", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("db --help should not error: %v", err)
	}
}
