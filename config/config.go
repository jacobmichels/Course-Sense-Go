package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Firestore struct {
		ProjectID           string `mapstructure:"project_id"`
		CredentialsFilePath string `mapstructure:"credentials_file"`
		SectionCollectionID string `mapstructure:"section_collection_id"`
		WatcherCollectionID string `mapstructure:"watcher_collection_id"`
	} `mapstructure:"firestore"`
	Smtp struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		From     string `mapstructure:"from"`
	} `mapstructure:"smtp"`
}

func ReadConfig() (Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("firestore.section_collection_id", "sections")
	viper.SetDefault("firestore.watcher_collection_id", "watchers")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, continuing with env and defaults")
		} else {
			// Config file was found but another error was produced
			return Config{}, fmt.Errorf("failed to read config file: %s", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %s", err)
	}

	return config, nil
}
