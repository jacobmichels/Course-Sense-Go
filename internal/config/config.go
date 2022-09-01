package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Firestore struct {
		ProjectID           string `mapstructure:"project_id"`
		CredentialsFilePath string `mapstructure:"credentials_file"`
	}
}

func ReadAppConfig() (*AppConfig, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("firestore.project_id", "")
	viper.SetDefault("firestore.credentials_file", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, continuing with env and defaults")
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read config file: %s", err)
		}
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return &config, nil
}
