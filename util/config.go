package util

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/spf13/viper"
)

// ConfigurationManager provides a method set for extracting data out of a configuration file.
type ConfigurationManager interface {
	GetInt(string) int
	GetBool(string) bool
	GetInt64(string) int64
	GetString(string) string

	SetConfigName(string)
	SetConfigType(string)
	SetEnvPrefix(string)
	SetEnvKeyReplacer(*strings.Replacer)

	AutomaticEnv()
	ReadInConfig() error
	AddConfigPath(string)
}

// Config represents the loaded configuration file.
var Config ConfigurationManager

// LoadConfig loads a configuration file specified by the function argument or a path provided the command line.
func LoadConfig(configFile string) ConfigurationManager {

	// NOTE: Alternate way for doing this: https://blog.gopheracademy.com/advent-2014/reading-config-files-the-go-way/

	var err error

	var configPath string
	var configName string

	var rootConfig ConfigurationManager

	if configFile == "" {
		configName = "config.yaml"
		configPath = os.Getenv("GOPATH") + "/src/gitlab.ido-services.com/luxtrust/base-component"
	} else {
		configPath, configName = filepath.Split(configFile)
	}

	rootConfig = viper.New()
	configName = strings.TrimSuffix(filepath.Ext(configName), ".yaml")

	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName(configName)
	rootConfig.AddConfigPath(configPath)
	rootConfig.SetEnvPrefix("cbc")
	rootConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	rootConfig.AutomaticEnv()

	if err = rootConfig.ReadInConfig(); err != nil {
		fmt.Printf("[error] configuration file could not be read: %v", err)
		os.Exit(1)
	}

	return rootConfig
}

func init() {
	// Initialization . . .
}
