package config

import (
	"log/slog"
	"slices"
	"time"
)

// validLogLevels is the set of default slog.Level values
// the slice is used to validate user log level input
var validLogLevels = []int{-4, 0, 4, 8}

type option func(*Config)

type Config struct {
	LogLevel     slog.Level
	LoggerOutput string
	Env          string
	ListenAddr   string
	Secret       string
	ConnStr      string
	TokenExpiry  time.Duration
}

func New(opts ...option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithServerOptions(env string, port string, maxPayloadSize int, connStr string) option {
	return func(c *Config) {
		c.Env = env
		c.ListenAddr = port
		c.ConnStr = connStr
	}
}

func WithLoggerOptions(logLevel int, loggerOutput string) option {
	if !slices.Contains(validLogLevels, logLevel) {
		panic("invalid log level value")
	}
	level := slog.Level(logLevel)
	return func(c *Config) {
		c.LogLevel = level
		c.LoggerOutput = loggerOutput
	}
}

func WithJWTOptions(secret string, tokenExpiry time.Duration) option {
	return func(c *Config) {
		c.Secret = secret
		c.TokenExpiry = tokenExpiry
	}
}
