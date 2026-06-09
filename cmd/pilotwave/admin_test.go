package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	auth_model "git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator/model"
	database "git.brobridge.com/pilotwave/pilotwave/pkg/database"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func TestResolveAdminPasswordFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(path, []byte("secret\n"), 0600); err != nil {
		t.Fatalf("failed to write password file: %v", err)
	}

	password, err := resolveAdminPassword("", path)
	if err != nil {
		t.Fatalf("resolveAdminPassword returned error: %v", err)
	}
	if password != "secret" {
		t.Fatalf("expected trimmed password from file, got %q", password)
	}
}

func TestResolveAdminPasswordRejectsMultipleSources(t *testing.T) {
	if _, err := resolveAdminPassword("secret", "password-file"); err == nil {
		t.Fatalf("expected multiple source error")
	}
}

func TestResolveAdminPasswordFromEnv(t *testing.T) {
	t.Setenv(adminPasswordEnv, "from-env")

	password, err := resolveAdminPassword("", "")
	if err != nil {
		t.Fatalf("resolveAdminPassword returned error: %v", err)
	}
	if password != "from-env" {
		t.Fatalf("expected env password, got %q", password)
	}
}

func TestResolveAdminPasswordRequiresPassword(t *testing.T) {
	t.Setenv(adminPasswordEnv, "")

	_, err := resolveAdminPassword("", "")
	if err == nil {
		t.Fatalf("expected missing password error")
	}
	if !strings.Contains(err.Error(), adminPasswordEnv) {
		t.Fatalf("expected error to mention env fallback, got %v", err)
	}
}

func TestHandleCLIIgnoresNonAdminCommand(t *testing.T) {
	handled, err := handleCLI([]string{"serve"})
	if err != nil {
		t.Fatalf("handleCLI returned error: %v", err)
	}
	if handled {
		t.Fatalf("expected non-admin command to be left for normal startup")
	}
}

func TestResetAdminPasswordUpdatesAdminFromPasswordFile(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	dbPath := filepath.Join(t.TempDir(), "pilotwave.db")
	passwordPath := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(passwordPath, []byte("changed-password\n"), 0600); err != nil {
		t.Fatalf("failed to write password file: %v", err)
	}

	viper.Set("database.type", "sqlite3")
	viper.Set("database.dbpath", dbPath)
	viper.Set("database.debug_mode", false)
	viper.Set("dev.reset_admin_password", false)

	if err := resetAdminPassword([]string{"--password-file", passwordPath}); err != nil {
		t.Fatalf("resetAdminPassword returned error: %v", err)
	}

	db := database.NewDatabase()
	if err := db.Init("sqlite3", dbPath); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	var user auth_model.User
	if err := db.GetInstance().Where("username = ?", "admin").Find(&user).Error; err != nil {
		t.Fatalf("failed to load admin user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("changed-password")); err != nil {
		t.Fatalf("password hash does not match password file: %v", err)
	}
	if user.IsDisabled {
		t.Fatalf("reset should enable the admin user")
	}
}

func TestResetAdminPasswordReturnsMissingUserError(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("database.type", "sqlite3")
	viper.Set("database.dbpath", filepath.Join(t.TempDir(), "pilotwave.db"))
	viper.Set("database.debug_mode", false)
	viper.Set("dev.reset_admin_password", false)

	err := resetAdminPassword([]string{"--username", "missing", "--password", "changed-password"})
	if err == nil {
		t.Fatalf("expected missing user error")
	}
	if !strings.Contains(err.Error(), `user "missing" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
