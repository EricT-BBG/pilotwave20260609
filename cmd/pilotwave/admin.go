package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	database "git.brobridge.com/pilotwave/pilotwave/pkg/database"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const adminPasswordEnv = "PILOTWAVE_ADMIN_PASSWORD"

func handleCLI(args []string) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	if args[0] != "admin" {
		return false, nil
	}

	if len(args) < 2 {
		return true, fmt.Errorf("missing admin command")
	}

	switch args[1] {
	case "reset-password":
		return true, resetAdminPassword(args[2:])
	default:
		return true, fmt.Errorf("unknown admin command %q", args[1])
	}
}

func resetAdminPassword(args []string) error {
	fs := flag.NewFlagSet("pilotwave admin reset-password", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	username := fs.String("username", "admin", "username to reset")
	password := fs.String("password", "", "new password; prefer --password-file or PILOTWAVE_ADMIN_PASSWORD")
	passwordFile := fs.String("password-file", "", "file containing the new password")

	if err := fs.Parse(args); err != nil {
		return err
	}

	resolvedPassword, err := resolveAdminPassword(*password, *passwordFile)
	if err != nil {
		return err
	}

	dbType := viper.GetString("database.type")
	uri, err := databaseURIFromConfig(dbType)
	if err != nil {
		return err
	}

	db := database.NewDatabase()
	if err := db.Init(dbType, uri); err != nil {
		return err
	}

	if err := db.ResetUserPassword(*username, resolvedPassword); err != nil {
		return err
	}

	log.Infof("Reset password for user %q", *username)
	return nil
}

func resolveAdminPassword(password string, passwordFile string) (string, error) {
	if password != "" && passwordFile != "" {
		return "", fmt.Errorf("use only one of --password or --password-file")
	}

	if passwordFile != "" {
		content, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		password = strings.TrimSpace(string(content))
	}

	if password == "" {
		password = os.Getenv(adminPasswordEnv)
	}

	if password == "" {
		return "", fmt.Errorf("password is required; use --password-file, --password, or %s", adminPasswordEnv)
	}

	return password, nil
}
