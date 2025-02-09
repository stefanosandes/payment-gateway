package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP           HTTPConfig           `mapstructure:"http"`
	Retry          RetryConfig          `mapstructure:"retry"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

type HTTPConfig struct {
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
}

type RetryConfig struct {
	Attempts     int `mapstructure:"attempts"`
	DelaySeconds int `mapstructure:"delay_seconds"`
}

type CircuitBreakerConfig struct {
	MaxRequests     uint32  `mapstructure:"max_requests"`
	IntervalSeconds int     `mapstructure:"interval_seconds"`
	TimeoutSeconds  int     `mapstructure:"timeout_seconds"`
	MinRequests     uint32  `mapstructure:"min_requests"`
	FailureRatio    float64 `mapstructure:"failure_ratio"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("http.timeout_seconds", 10)
	viper.SetDefault("retry.attempts", 3)
	viper.SetDefault("retry.delay_seconds", 1)
	viper.SetDefault("circuit_breaker.max_requests", 3)
	viper.SetDefault("circuit_breaker.interval_seconds", 10)
	viper.SetDefault("circuit_breaker.timeout_seconds", 30)
	viper.SetDefault("circuit_breaker.min_requests", 3)
	viper.SetDefault("circuit_breaker.failure_ratio", 0.6)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func (c *Config) GetHTTPTimeout() time.Duration {
	return time.Duration(c.HTTP.TimeoutSeconds) * time.Second
}

func (c *Config) GetRetryDelay() time.Duration {
	return time.Duration(c.Retry.DelaySeconds) * time.Second
}

func (c *Config) GetCircuitBreakerInterval() time.Duration {
	return time.Duration(c.CircuitBreaker.IntervalSeconds) * time.Second
}

func (c *Config) GetCircuitBreakerTimeout() time.Duration {
	return time.Duration(c.CircuitBreaker.TimeoutSeconds) * time.Second
}
