package config

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"sync"
)

type Config struct {
	RedisAddr                 string `mapstructure:"REDIS_ADDR"`
	Port                      string `mapstructure:"PORT"`
	RedisPassword             string `mapstructure:"REDIS_PASSWORD"`
	CircuitBreakerName        string `mapstructure:"CIRCUIT_BREAKER_NAME"`
	CircuitBreakerMaxRequests int    `mapstructure:"CIRCUIT_BREAKER_MAX_REQUESTS"`
	CircuitBreakerInterval    int    `mapstructure:"CIRCUIT_BREAKER_INTERVAL"`
	CircuitBreakerTimeout     int    `mapstructure:"CIRCUIT_BREAKER_TIMEOUT"`
	TLSMinVersion             uint16 `mapstructure:"TLS_MIN_VERSION"`
}

var (
	cachedConfig *Config
	once         sync.Once
)

func LoadConfig() (*Config, error) {
	var err error

	once.Do(func() {
		viper.AutomaticEnv()
		viper.SetDefault("REDIS_ADDR", "redis:6379")
		viper.SetDefault("PORT", "8080")
		viper.SetDefault("REDIS_PASSWORD", "")
		viper.SetDefault("CIRCUIT_BREAKER_NAME", "RedisCircuitBreaker")
		viper.SetDefault("CIRCUIT_BREAKER_MAX_REQUESTS", 5)
		viper.SetDefault("CIRCUIT_BREAKER_INTERVAL", 60)
		viper.SetDefault("CIRCUIT_BREAKER_TIMEOUT", 30)
		viper.SetDefault("TLS_MIN_VERSION", tls.VersionTLS12)

		var cfg Config
		if err = viper.Unmarshal(&cfg); err != nil {
			err = errors.Wrap(err, "failed to unmarshal config")
			return
		}
		cachedConfig = &cfg
	})

	if cachedConfig == nil {
		return nil, err
	}
	return cachedConfig, nil
}
