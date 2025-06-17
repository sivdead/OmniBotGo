package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type (
	// Config -.
	Config struct {
		App     App     `mapstructure:"app"`
		HTTP    HTTP    `mapstructure:"http"`
		Log     Log     `mapstructure:"log"`
		DB      DB      `mapstructure:"db"`
		GRPC    GRPC    `mapstructure:"grpc"`
		RMQ     RMQ     `mapstructure:"rmq"`
		Metrics Metrics `mapstructure:"metrics"`
		Swagger Swagger `mapstructure:"swagger"`
	}

	// App -.
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	}

	// HTTP -.
	HTTP struct {
		Port           string `mapstructure:"port"`
		UsePreforkMode bool   `mapstructure:"use_prefork_mode"`
	}

	// Log -.
	Log struct {
		Level string `mapstructure:"level"`
	}

	// DB -.
	DB struct {
		Type           string `mapstructure:"type"`
		DSN            string `mapstructure:"dsn"`
		MaxConnections int    `mapstructure:"max_connections"`
		LogLevel       string `mapstructure:"log_level"`
	}

	// GRPC -.
	GRPC struct {
		Port string `mapstructure:"port"`
	}

	// RMQ -.
	RMQ struct {
		ServerExchange string `mapstructure:"server_exchange"`
		ClientExchange string `mapstructure:"client_exchange"`
		URL            string `mapstructure:"url"`
	}

	// Metrics -.
	Metrics struct {
		Enabled bool `mapstructure:"enabled"`
	}

	// Swagger -.
	Swagger struct {
		Enabled bool `mapstructure:"enabled"`
	}
)

// NewConfig returns app config using Viper.
// It supports multiple configuration sources with the following priority:
// 1. Explicit calls to Set
// 2. Environment variables
// 3. Config file
// 4. Defaults
func NewConfig() (*Config, error) {
	// Create new viper instance
	v := viper.New()

	// Set configuration file name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")        // current directory
	v.AddConfigPath("./config") // config subdirectory

	// Set defaults
	setDefaults(v)

	// Enable reading from environment variables
	v.AutomaticEnv()

	// Set environment variable key replacer
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set environment variable prefix
	v.SetEnvPrefix("") // No prefix to keep compatibility

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Warning: Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
	}

	// Unmarshal config into struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "OmniBotGo")
	v.SetDefault("app.version", "1.0.0")

	// HTTP defaults
	v.SetDefault("http.port", "8080")
	v.SetDefault("http.use_prefork_mode", false)

	// Log defaults
	v.SetDefault("log.level", "info")

	// DB defaults
	v.SetDefault("db.type", "mysql")
	v.SetDefault("db.max_connections", 10)
	v.SetDefault("db.log_level", "warn")

	// gRPC defaults
	v.SetDefault("grpc.port", "8081")

	// RMQ defaults
	v.SetDefault("rmq.server_exchange", "rpc_server")
	v.SetDefault("rmq.client_exchange", "rpc_client")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)

	// Swagger defaults
	v.SetDefault("swagger.enabled", false)
}

// validateConfig validates required configuration fields
func validateConfig(cfg *Config) error {
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if cfg.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}
	if cfg.HTTP.Port == "" {
		return fmt.Errorf("http.port is required")
	}
	if cfg.Log.Level == "" {
		return fmt.Errorf("log.level is required")
	}
	if cfg.DB.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	if cfg.GRPC.Port == "" {
		return fmt.Errorf("grpc.port is required")
	}
	if cfg.RMQ.ServerExchange == "" {
		return fmt.Errorf("rmq.server_exchange is required")
	}
	if cfg.RMQ.ClientExchange == "" {
		return fmt.Errorf("rmq.client_exchange is required")
	}
	if cfg.RMQ.URL == "" {
		return fmt.Errorf("rmq.url is required")
	}
	return nil
}

// GetGlobalViper returns a global viper instance for migration and other uses
var globalViper *viper.Viper

func init() {
	// Initialize global viper instance
	globalViper = viper.New()

	// Set configuration file name and paths
	globalViper.SetConfigName("config")
	globalViper.SetConfigType("yaml")
	globalViper.AddConfigPath(".")
	globalViper.AddConfigPath("./config")

	// Set defaults
	setDefaults(globalViper)

	// Enable reading from environment variables
	globalViper.AutomaticEnv()
	globalViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file
	if err := globalViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}

// GetString returns string value from global viper
func GetString(key string) string {
	return globalViper.GetString(key)
}

// GetInt returns int value from global viper
func GetInt(key string) int {
	return globalViper.GetInt(key)
}

// GetBool returns bool value from global viper
func GetBool(key string) bool {
	return globalViper.GetBool(key)
}
