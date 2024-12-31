package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Server struct {
		Port int
	}
	Plaid struct {
		ClientID     string
		Secret       string
		Environment  string
		RedirectURI  string
		WebhookURL   string
	}
	JWT struct {
		Secret string
	}
	Email struct {
		SMTPHost     string
		SMTPPort     int
		SMTPUser     string
		SMTPPassword string
		FromAddress  string
	}
}

var cfg *Config

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	cfg = &Config{}

	// Database configuration
	cfg.Database.Host = viper.GetString("DATABASE_HOST")
	cfg.Database.Port = viper.GetInt("DATABASE_PORT")
	cfg.Database.User = viper.GetString("DATABASE_USER")
	cfg.Database.Password = viper.GetString("DATABASE_PASSWORD")
	cfg.Database.Name = viper.GetString("DATABASE_NAME")
	cfg.Database.SSLMode = viper.GetString("DATABASE_SSLMODE")

	// Server configuration
	cfg.Server.Port = viper.GetInt("SERVER_PORT")

	// Plaid configuration
	cfg.Plaid.ClientID = viper.GetString("PLAID_CLIENT_ID")
	cfg.Plaid.Secret = viper.GetString("PLAID_SECRET")
	cfg.Plaid.Environment = viper.GetString("PLAID_ENVIRONMENT")
	cfg.Plaid.RedirectURI = viper.GetString("PLAID_REDIRECT_URI")
	cfg.Plaid.WebhookURL = viper.GetString("PLAID_WEBHOOK_URL")

	// JWT configuration
	cfg.JWT.Secret = viper.GetString("JWT_SECRET")

	// Email configuration
	cfg.Email.SMTPHost = viper.GetString("EMAIL_SMTP_HOST")
	cfg.Email.SMTPPort = viper.GetInt("EMAIL_SMTP_PORT")
	cfg.Email.SMTPUser = viper.GetString("EMAIL_SMTP_USER")
	cfg.Email.SMTPPassword = viper.GetString("EMAIL_SMTP_PASSWORD")
	cfg.Email.FromAddress = viper.GetString("EMAIL_FROM_ADDRESS")

	return cfg, nil
}

func Get() *Config {
	return cfg
}
