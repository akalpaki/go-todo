package config

import "time"

type option func(*Configv2)

type Configv2 struct {
	logLevel       int
	maxPayloadSize int
	loggerOutput   string
	env            string
	listenAddr     string
	connStr        string
	secret         string
	tokenExpiry    time.Duration
}

func New(opts ...option) *Configv2 {
	cfg := &Configv2{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithServerOptions(env string, port string, maxPayloadSize int, connStr string) option {
	return func(c *Configv2) {
		c.env = env
		c.listenAddr = port
		c.maxPayloadSize = maxPayloadSize
		c.connStr = connStr // TODO: create function which groups db based dependencies together when implementing pgSQL
	}
}

func WithLoggerOptions(logLevel int, loggerOutput string) option {
	return func(c *Configv2) {
		c.logLevel = logLevel
		c.loggerOutput = loggerOutput
	}
}

func WithJWTOptions(secret string, tokenExpiry time.Duration) option {
	return func(c *Configv2) {
		c.secret = secret
		c.tokenExpiry = tokenExpiry
	}
}
