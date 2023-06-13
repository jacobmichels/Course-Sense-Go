package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// returns a slice of the supported db names
// a function is used to access this list instead of a global slice to prevent accidental mutation of the slice
func getSupportedDbs() []string {
	return []string{"sqlite", "firestore"}
}

// reads the config file
// returns an error if the file couldn't be read or is invalid
func ParseConfig() (Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetDefault("poll_interval_secs", 300)
	viper.SetDefault("database.type", "")
	viper.SetDefault("database.firestore.project_id", "")
	viper.SetDefault("database.firestore.credentials_file", "")
	viper.SetDefault("database.firestore.section_collection_id", "sections")
	viper.SetDefault("database.firestore.watcher_collection_id", "watchers")
	viper.SetDefault("database.sqlite.connection_string", "")
	viper.SetDefault("notifications.emailsmtp.port", 0)
	viper.SetDefault("notifications.emailsmtp.host", "")
	viper.SetDefault("notifications.emailsmtp.username", "")
	viper.SetDefault("notifications.emailsmtp.password", "")
	viper.SetDefault("notifications.emailsmtp.from", "")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, continuing with env and defaults")
		} else {
			// Config file was found but another error was produced
			return Config{}, fmt.Errorf("failed to read config file: %s", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	if cfg.Database.Type == "" {
		return fmt.Errorf("no database type set. database type can be one of: %v", getSupportedDbs())
	} else if cfg.Database.Type != "sqlite" && cfg.Database.Type != "firestore" {
		return fmt.Errorf("bad database type. database type can be one of: %v", getSupportedDbs())
	}

	if cfg.Database.Type == "sqlite" && cfg.Database.SQLite.ConnectionString == "" {
		log.Printf("warn: sqlite connection string is empty")
	}

	return nil
}
