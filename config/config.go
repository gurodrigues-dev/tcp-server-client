package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Environments struct {
	ServerAddr       string `mapstructure:"SERVER_ADDR"`
	DatabaseHost     string `mapstructure:"DATABASE_HOST"`
	DatabasePort     string `mapstructure:"DATABASE_PORT"`
	DatabaseUser     string `mapstructure:"DATABASE_USER"`
	DatabaseName     string `mapstructure:"DATABASE_NAME"`
	DatabasePassword string `mapstructure:"DATABASE_PASSWORD"`
	DatabaseSSLMode  string `mapstructure:"DATABASE_SSLMODE"`
	RabbitMQHost     string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQPort     string `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUser     string `mapstructure:"RABBITMQ_USER"`
	RabbitMQPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitMQVHost    string `mapstructure:"RABBITMQ_VHOST"`
	RabbitMQExchange string `mapstructure:"RABBITMQ_EXCHANGE"`
	RabbitMQType     string `mapstructure:"RABBITMQ_TYPE"`
	RabbitMQTopic    string `mapstructure:"RABBITMQ_TOPIC"`
	LogDir           string `mapstructure:"LOG_DIR"`
}

func Load() (*Environments, error) {
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("fail to read .env: %w", err)
		}
		viper.AutomaticEnv()

		env := &Environments{}
		if err := viper.Unmarshal(env); err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}
		return env, nil
	}

	return nil, fmt.Errorf("no .env file found")
}
