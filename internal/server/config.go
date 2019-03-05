package server

import (
	"github.com/namsral/flag"
)

// Configuration holds all of the service's configuration keys
type Configuration struct {
	Port int

	MessageBirdKey string

	CircuitBreakerTimeoutSeconds        int
	CircuitBreakerSleepWindowSeconds    int
	CircuitBreakerErrorPercentThreshold int

	RateLimitMaxRequests      int
	RateLimitPerPeriodSeconds int

	RedisHost               string
	RedisPwd                string
	RedisMaxIdle            int
	RedisMaxActive          int
	RedisIdleTimeoutSeconds int

	BufferDBConnectionString string
	BufferDBMaxConnections   int

	LogFormat string
}

// InitConfig loads configuration from env variables
func InitConfig() Configuration {
	cfg := Configuration{}

	flag.IntVar(&cfg.Port, "listen_port", 8085, "The port for the server to listen on")
	flag.StringVar(&cfg.LogFormat, "log_format", "text", "Logger format: can be 'text' or 'json'")

	flag.StringVar(&cfg.MessageBirdKey, "message_bird_key", "", "Logger format: can be 'text' or 'json'")

	flag.IntVar(&cfg.CircuitBreakerTimeoutSeconds, "circuit_breaker_timeout", 120, "The port for the server to listen on")
	flag.IntVar(&cfg.CircuitBreakerSleepWindowSeconds, "circuit_breaker_sleep_window", 10, "The port for the server to listen on")
	flag.IntVar(&cfg.CircuitBreakerErrorPercentThreshold, "circuit_breaker_error_percent_threshold", 10, "The port for the server to listen on")

	flag.IntVar(&cfg.RateLimitMaxRequests, "rate_limit_max_requests", 100, "The port for the server to listen on")
	flag.IntVar(&cfg.RateLimitPerPeriodSeconds, "rate_limit_per_period", 1, "The port for the server to listen on")

	flag.IntVar(&cfg.BufferDBMaxConnections, "buffer_db_max_conns", 5, "The port for the server to listen on")
	flag.StringVar(&cfg.BufferDBConnectionString, "buffer_db_connection_string", "", "Logger format: can be 'text' or 'json'")

	flag.StringVar(&cfg.RedisHost, "redis_host", ":6379", "The port for the server to listen on")
	flag.StringVar(&cfg.RedisPwd, "redis_pwd", "", "The port for the server to listen on")
	flag.IntVar(&cfg.RedisMaxIdle, "redis_max_idle", 20, "The port for the server to listen on")
	flag.IntVar(&cfg.RedisMaxActive, "redis_max_active", 20, "The port for the server to listen on")
	flag.IntVar(&cfg.RedisIdleTimeoutSeconds, "redis_max_idle_timeout", 240, "The port for the server to listen on")

	flag.Parse()

	return cfg
}
