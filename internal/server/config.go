package server

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
		CollectionID        string `mapstructure:"collection_id"`
	}
	Twilio struct {
		AccountSID  string `mapstructure:"account_sid"`
		AuthToken   string `mapstructure:"auth_token"`
		PhoneNumber string `mapstructure:"phone_number"`
	}
	Smtp struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		From     string `mapstructure:"from"`
	}
}

func readAppConfig() (*AppConfig, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("firestore.project_id", "")
	viper.SetDefault("firestore.credentials_file", "")
	viper.SetDefault("firestore.collection_id", "sections")

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
