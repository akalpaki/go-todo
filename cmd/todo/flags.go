package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/akalpaki/todo/internal/config"
)

const (
	defaulLogLevel     = -4 // debug level in log/slog
	defualtTokenExpiry = 30 * time.Minute
)

var (
	env            string
	listenAddr     string
	maxPayloadSize int
	connStr        string
	dbName         string
	dbPort         string
	dbUser         string
	dbPass         string
	logLevel       int
	loggerOutput   string
	secret         string
	tokenExpiry    time.Duration
	h              bool
)

func init() {
	flag.StringVar(&env, "env", lookupEnvString("ENV", "dev"), "the name of the environment the server is being run")
	flag.StringVar(&listenAddr, "port", lookupEnvString("PORT", ":8000"), "the port the server is listening at")
	flag.StringVar(&connStr, "conn_str", lookupEnvString("CONNECTION_STRING", "file:todo.db"), "database connection string")
	flag.IntVar(&logLevel, "log_level", lookupEnvInt("LOG_LEVEL", defaulLogLevel), "minimum logging level")
	flag.StringVar(&loggerOutput, "log_output", lookupEnvString("LOG_OUTPUT", os.Stdout.Name()), "path to the logger's output file")
	flag.StringVar(&secret, "secret", lookupEnvString("JWT_SECRET_KEY", "secret"), "jwt signing key")
	flag.DurationVar(&tokenExpiry, "token_exp", lookupEnvDuration("TOKEN_EXPIRY", defualtTokenExpiry), "expiration time of jwt tokens")
	flag.BoolVar(&h, "h", false, "prints help text")

	flag.Parse()

	if h {
		help()
	}
}

func loadConfig() *config.Config {
	log.Printf("loading configuration for %s environment", env)
	return config.New(
		config.WithServerOptions(
			env,
			listenAddr,
			maxPayloadSize,
			connStr,
		),
		config.WithLoggerOptions(
			logLevel,
			loggerOutput,
		),
		config.WithJWTOptions(
			secret,
			tokenExpiry,
		),
	)
}

// lookupEnvXXX functions provide fallback config values for unspecified flags.
// They attempt to look up env vars, providing sensible defaults if not found

func lookupEnvString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("failed to read integer environment variable %s, error=%s", key, err.Error())
		}
		return intVal
	}
	return defaultVal
}

func lookupEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		durVal, err := time.ParseDuration(val)
		if err != nil {
			log.Fatalf("failed to read time duration environment variable %s, error=%s", key, err.Error())
		}
		return durVal
	}
	return defaultVal
}

func help() {
	text := `
	__________  ____  ____ 
	/_  __/ __ \/ __ \/ __ \
	 / / / / / / / / / / / /
	/ / / /_/ / /_/ / /_/ / 
       /_/  \____/_____/\____/  

	go-todo is a minimalistic todo REST api written in go.

	The following configuration options can be provided through CLI flags:
	--env : the name of the environment the server is being run on (eg. prod)
	--port : the port number (should be between 1024 and 65535)
	--log_level : the minimum logging level (should be one of the values listed below)
		Available values: -4 (debug), 0 (info), 4 (warn), 8 (error)
	--log_output : the path to the logger's output file
		default :  stdout 
	--secret : jwt secret key, used in signing and validating jwt tokens
	--token_exp : duration of jwt validity
		ATTENTION: this should be formatted as a string that can be parsed by time.ParseDuration (https://pkg.go.dev/time#ParseDuration)
		default :  30 minutes
	--conn_str : database connection string
	`
	fmt.Println(text)
	os.Exit(0)
}
