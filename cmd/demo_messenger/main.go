package main

import (
	"github.com/arkadyb/caply"
	"github.com/arkadyb/demo_messenger/internal/messenger"
	"github.com/arkadyb/demo_messenger/internal/pkg/buffer"
	"github.com/arkadyb/demo_messenger/internal/server"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"net/http"
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
		cfg = server.InitConfig()
		cp  *caply.Caply
		err error
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

	// create twilio http client
	httpclient := &http.Client{}

	// init application
	messenger := messenger.NewMessenger(messenger.SendSMSViaTwilio(cfg.TwilioSid, cfg.TwilioToken, httpclient), buffer)
	messenger.Errors = make(chan error)
	go func() {
		for err := range messenger.Errors {
			log.Error(err)
		}
	}()

	// init redis db and rate-limiter
	rateLimiterStore := caply.NewRedisStore(&redis.Pool{
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
	cp, err = caply.NewCaply(cfg.RateLimitMaxRequests, time.Duration(cfg.RateLimitPerPeriodSeconds)*time.Second, rateLimiterStore)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "failed to setup rate limiter"))
	}

	server := server.NewServer(cfg, messenger, cp)
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
