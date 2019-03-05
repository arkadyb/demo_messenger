package main

import (
	"github.com/arkadyb/demo_messenger/internal/messenger"
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/arkadyb/demo_messenger/internal/server"
	"github.com/arkadyb/rate_limiter"
	"github.com/gomodule/redigo/redis"
	"github.com/messagebird/go-rest-api"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	version    = "0.0.0-dev"
	creator    = "dev"
	buildstamp = "today"
)

func main() {
	var (
		cfg         = server.InitConfig()
		rateLimiter ratelimiter.RateLimiter
		err         error
	)

	// Setup log format
	if cfg.LogFormat == strings.ToLower("json") {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Log version details
	log.WithFields(log.Fields{
		"version":    version,
		"author":     creator,
		"buildstamp": buildstamp,
	}).Info("build information")

	// workaround for demo purposes; docker starts postgres composer quite fast, but postgres itself is not ready to accept connections at the time
	time.Sleep(5 * time.Second)
	// init buffer store for queue of messages waiting to be delivered
	buffer, err := buffer.NewPostgresBuffer(cfg.BufferDBConnectionString, cfg.BufferDBMaxConnections)
	if err != nil {
		log.Fatalln("failed to dial postgres:", err)
	}

	// init application
	messenger := messenger.NewMessenger(messenger.SendSMSViaMessageBird(messagebird.New(cfg.MessageBirdKey)), buffer)
	messenger.Errors = make(chan error)
	go func() {
		for err := range messenger.Errors {
			log.Error(err)
		}
	}()

	// init redis db and rate-limiter
	rateLimiterStore := ratelimiter.NewRedisStore(&redis.Pool{
		MaxIdle:     cfg.RedisMaxIdle,
		MaxActive:   cfg.RedisMaxActive,
		IdleTimeout: time.Duration(cfg.RedisIdleTimeoutSeconds) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.RedisHost, redis.DialPassword(cfg.RedisPwd))
			if err != nil {
				log.Fatalln("failed to dial redis:", err)
			}
			return c, err
		},
	})
	rateLimiter, err = ratelimiter.NewFixedTimeWindowRateLimiter(cfg.RateLimitMaxRequests, time.Duration(cfg.RateLimitPerPeriodSeconds)*time.Second, rateLimiterStore)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to setup rate limiter"))
	}

	server := server.NewServer(cfg, messenger, rateLimiter)
	// start server
	server.Start()

	// wait for SIGTERM or SIGINT signals to shutdown app and server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	signal := <-c

	// resources cleanup
	messenger.Shutdown()
	buffer.Close()
	server.Stop()
	log.Fatalf("process killed with signal: %v", signal.String())
}
