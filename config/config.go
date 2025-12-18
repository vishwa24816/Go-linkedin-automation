package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application's configuration settings.
type Config struct {
	LinkedIn struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"linkedin"`
	// Add other configuration fields here as needed
}

// LoadConfig reads configuration from file and environment variables.
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // or "json"
	viper.AddConfigPath(".")      // path to look for the config file in the current directory
	viper.AddConfigPath("./config") // path to look for the config file in the config directory

	// Read environment variables
	viper.SetEnvPrefix("LINKEDIN_AUTOMATION") // prefix for environment variables
	viper.AutomaticEnv()                    // read in environment variables that match

	// Set default values
	viper.SetDefault("linkedin.username", "")
	viper.SetDefault("linkedin.password", "")

	var cfg Config

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if we're relying solely on environment variables
			fmt.Println("Config file not found, relying on environment variables or defaults.")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate essential configuration
	if cfg.LinkedIn.Username == "" || cfg.LinkedIn.Password == "" {
		return nil, fmt.Errorf("linkedin username and password must be provided (either in config file or via environment variables LINKEDIN_AUTOMATION_LINKEDIN_USERNAME and LINKEDIN_AUTOMATION_LINKEDIN_PASSWORD)")
	}

	return &cfg, nil
}

// SaveConfig writes the current configuration to a file (optional, for persistent changes)
func SaveConfig(cfg *Config, filePath string) error {
	viper.Set("linkedin.username", cfg.LinkedIn.Username)
	viper.Set("linkedin.password", cfg.LinkedIn.Password)
	// Set other fields if needed

	// Ensure the directory exists
	// For simplicity, we'll just write to the root for now,
	// but in a real app, you'd want to handle paths properly.
	return viper.WriteConfigAs(filePath)
}
