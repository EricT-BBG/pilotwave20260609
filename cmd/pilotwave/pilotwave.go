package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	app "git.brobridge.com/pilotwave/pilotwave/pkg/app/instance"
	"git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type globalOptions struct {
	ConfigPath  string
	ShowVersion bool
	Args        []string
}

func configure(configPath string) error {

	// From the environment
	viper.SetEnvPrefix("PILOTWAVE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// From config file
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./configs")

	if configPath != "" {
		viper.SetConfigFile(configPath)
		if strings.HasSuffix(configPath, ".toml.dist") {
			viper.SetConfigType("toml")
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		if configPath != "" {
			return fmt.Errorf("failed to load config file %q: %w", configPath, err)
		}
		log.Warn("No configuration file was loaded")
	}

	return nil
}

func parseGlobalOptions(args []string) (globalOptions, error) {
	fs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "", "path to config.toml")
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return globalOptions{}, err
	}

	return globalOptions{
		ConfigPath:  *configPath,
		ShowVersion: *showVersion,
		Args:        fs.Args(),
	}, nil
}

//go:generate go generate ../../pkg/http_server/static
func main() {
	opts, err := parseGlobalOptions(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if err := configure(opts.ConfigPath); err != nil {
		log.Fatal(err)
	}

	if opts.ShowVersion {
		version, commit, buildTime := buildinfo.Values()
		fmt.Printf("Pilotwave %s (commit %s, built %s)\n", version, commit, buildTime)
		return
	}

	handled, err := handleCLI(opts.Args)
	if handled {
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Initializing application
	a := app.NewAppInstance()

	err = a.Init()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Starting application
	err = a.Run()
	if err != nil {
		log.Fatal(err)
		return
	}
}
