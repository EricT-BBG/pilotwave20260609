package database

import (
	"path/filepath"
	"testing"

	auth_model "git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator/model"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func configureDatabaseTest(t *testing.T, resetAdminPassword bool) {
	t.Helper()

	viper.Reset()
	viper.Set("database.debug_mode", false)
	viper.Set("dev.reset_admin_password", resetAdminPassword)
	t.Cleanup(viper.Reset)
}

func TestResetUserPasswordUpdatesExistingUser(t *testing.T) {
	configureDatabaseTest(t, false)

	db := NewDatabase()
	dbPath := filepath.Join(t.TempDir(), "pilotwave.db")
	if err := db.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if err := db.ResetUserPassword("admin", "changed-password"); err != nil {
		t.Fatalf("ResetUserPassword returned error: %v", err)
	}

	var user auth_model.User
	if err := db.db.Where("username = ?", "admin").Find(&user).Error; err != nil {
		t.Fatalf("failed to load admin user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("changed-password")); err != nil {
		t.Fatalf("password hash does not match changed password: %v", err)
	}
	if user.IsDisabled {
		t.Fatalf("reset should enable the user")
	}
}

func TestResetUserPasswordRequiresExistingUser(t *testing.T) {
	configureDatabaseTest(t, false)

	db := NewDatabase()
	dbPath := filepath.Join(t.TempDir(), "pilotwave.db")
	if err := db.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	if err := db.ResetUserPassword("missing", "changed-password"); err == nil {
		t.Fatalf("expected missing user error")
	}
}

func TestInitDoesNotResetAdminPasswordByDefault(t *testing.T) {
	configureDatabaseTest(t, false)

	dbPath := filepath.Join(t.TempDir(), "pilotwave.db")
	db := NewDatabase()
	if err := db.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := db.ResetUserPassword("admin", "changed-password"); err != nil {
		t.Fatalf("ResetUserPassword returned error: %v", err)
	}
	if err := db.db.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	reopened := NewDatabase()
	if err := reopened.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("reopened Init returned error: %v", err)
	}

	user := loadAdminUser(t, reopened)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("changed-password")); err != nil {
		t.Fatalf("admin password should not reset without dev flag: %v", err)
	}
}

func TestInitResetsAdminPasswordWhenDevFlagEnabled(t *testing.T) {
	configureDatabaseTest(t, false)

	dbPath := filepath.Join(t.TempDir(), "pilotwave.db")
	db := NewDatabase()
	if err := db.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := db.ResetUserPassword("admin", "changed-password"); err != nil {
		t.Fatalf("ResetUserPassword returned error: %v", err)
	}
	if err := db.db.Model(&auth_model.User{}).Where("username = ?", "admin").Updates(map[string]interface{}{
		"Permissions": "viewer",
		"IsDisabled":  true,
	}).Error; err != nil {
		t.Fatalf("failed to alter admin user: %v", err)
	}
	if err := db.db.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	viper.Set("dev.reset_admin_password", true)
	reopened := NewDatabase()
	if err := reopened.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("reopened Init returned error: %v", err)
	}

	user := loadAdminUser(t, reopened)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("admin")); err != nil {
		t.Fatalf("admin password should reset to admin when dev flag is enabled: %v", err)
	}
	if user.Permissions != "admin" {
		t.Fatalf("expected admin permissions, got %q", user.Permissions)
	}
	if user.IsDisabled {
		t.Fatalf("admin should be enabled after dev reset")
	}
}

func loadAdminUser(t *testing.T, db *Database) auth_model.User {
	t.Helper()

	var user auth_model.User
	if err := db.db.Where("username = ?", "admin").Find(&user).Error; err != nil {
		t.Fatalf("failed to load admin user: %v", err)
	}
	return user
}
