package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv         string `mapstructure:"APP_ENV"`
	TursoURL       string `mapstructure:"TURSO_URL"`
	TursoAuthToken string `mapstructure:"TURSO_AUTH_TOKEN"`
	ServerPort     string `mapstructure:"PORT"` // Changed from SERVER_PORT to PORT for Render
	JWTSecret      string `mapstructure:"JWT_SECRET"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv() // Read from environment variables if set

	// Explicitly bind environment variables so Viper knows to look for them even without a .env file
	viper.BindEnv("APP_ENV")
	viper.BindEnv("TURSO_URL")
	viper.BindEnv("TURSO_AUTH_TOKEN")
	viper.BindEnv("PORT")
	viper.BindEnv("JWT_SECRET")

	// Read from .env file if it exists (mostly for local development)
	if err := viper.ReadInConfig(); err != nil {
		// It's ok if .env is missing in production, we'll use system environment variables
		if !os.IsNotExist(err) && viper.GetString("APP_ENV") == "development" {
			fmt.Printf("Warning: error reading .env file: %v\n", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}

	// Set defaults
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.AppEnv == "" {
		config.AppEnv = "development"
	}

	return &config, nil
}
