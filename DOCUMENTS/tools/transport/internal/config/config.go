package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName        string
	DBDebug       bool
	Hostname      string
	Host          string
	Cache         time.Duration
	LocalDelivery string
	Delivery      string
}

func LoadConfig(configPath string) Config {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.user", "postfixadmin")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "postfix")
	viper.SetDefault("database.debug", false)
	viper.SetDefault("transport.hostname", "localhost")
	viper.SetDefault("transport.host", "127.0.0.1:12221")
	viper.SetDefault("transport.cache", "10m")
	viper.SetDefault("transport.localdelivery", "")
	viper.SetDefault("transport.delivery", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read '%s'. Using default values: %v", configPath, err)
	}

	return Config{
		DBHost:   viper.GetString("database.host"),
		DBPort:   viper.GetString("database.port"),
		DBUser:   viper.GetString("database.user"),
		DBPass:        viper.GetString("database.password"),
		DBName:        viper.GetString("database.name"),
		DBDebug:       viper.GetBool("database.debug"),
		Hostname:      viper.GetString("transport.hostname"),
		Host:          viper.GetString("transport.host"),
		Cache:         viper.GetDuration("transport.cache"),
		LocalDelivery: viper.GetString("transport.localdelivery"),
		Delivery:      viper.GetString("transport.delivery"),
	}
}
