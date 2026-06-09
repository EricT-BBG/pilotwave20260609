package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestParseGlobalOptionsConfig(t *testing.T) {
	opts, err := parseGlobalOptions([]string{"--config", "/tmp/config.toml", "admin", "reset-password"})
	if err != nil {
		t.Fatalf("parseGlobalOptions returned error: %v", err)
	}

	if opts.ConfigPath != "/tmp/config.toml" {
		t.Fatalf("expected config path, got %q", opts.ConfigPath)
	}
	if len(opts.Args) != 2 || opts.Args[0] != "admin" || opts.Args[1] != "reset-password" {
		t.Fatalf("unexpected remaining args: %#v", opts.Args)
	}
}

func TestParseGlobalOptionsVersion(t *testing.T) {
	opts, err := parseGlobalOptions([]string{"--version"})
	if err != nil {
		t.Fatalf("parseGlobalOptions returned error: %v", err)
	}

	if !opts.ShowVersion {
		t.Fatalf("expected version flag")
	}
	if len(opts.Args) != 0 {
		t.Fatalf("unexpected remaining args: %#v", opts.Args)
	}
}

func TestParseGlobalOptionsRejectsUnknownFlag(t *testing.T) {
	if _, err := parseGlobalOptions([]string{"--missing"}); err == nil {
		t.Fatalf("expected unknown flag error")
	}
}

func TestConfigureReadsTomlDistConfig(t *testing.T) {
	viper.Reset()

	configPath := filepath.Join("..", "..", "configs", "config.toml.dist")
	if err := configure(configPath); err != nil {
		t.Fatalf("configure returned error: %v", err)
	}

	if got := viper.GetString("database.type"); got != "sqlite3" {
		t.Fatalf("expected sqlite3 database type, got %q", got)
	}
}

func TestConfigureReturnsErrorForMissingExplicitConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	err := configure(filepath.Join(t.TempDir(), "missing.toml"))
	if err == nil {
		t.Fatalf("expected missing explicit config error")
	}
	if !strings.Contains(err.Error(), "failed to load config file") {
		t.Fatalf("unexpected error: %v", err)
	}
}
