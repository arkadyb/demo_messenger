package server

import (
	"context"
	"fmt"
	hrx "github.com/afex/hystrix-go/hystrix"
	"github.com/arkadyb/caply"
	"github.com/arkadyb/demo_messenger/internal/messenger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

// NewServer returns new server instance
func NewServer(cfg Configuration, messenger messenger.Application, caply *caply.Caply) *Server {
	var (
		addr             = fmt.Sprintf(":%s", strconv.Itoa(cfg.Port))
		hrxDefaultConfig = hrx.CommandConfig{
			Timeout:               cfg.CircuitBreakerTimeoutSeconds * 1000,
			SleepWindow:           cfg.CircuitBreakerSleepWindowSeconds * 1000,
			ErrorPercentThreshold: cfg.CircuitBreakerErrorPercentThreshold,
		}
	)

	router := mux.NewRouter()
	router.Use(LoggingMiddleware, RateLimitingMiddleware(caply), PrometheusMiddleware, handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	router.NotFoundHandler = http.HandlerFunc(notFound404Handler)
	router.Handle("/health", http.HandlerFunc(healthHandler)).Methods("GET")
	router.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})).Methods("GET")

	v1 := router.PathPrefix("/v1/send").Subrouter()
	v1.Handle("/sms", CircuitBreakerMiddleware("ping_request", hrxDefaultConfig, SendSMSHandler(messenger))).Methods("POST")

	return &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

// Http server
type Server struct {
	*http.Server
}

// Stop gracefully stops server
func (s *Server) Stop() {
	log.Println("gracefully stopping server...")

	err := s.Shutdown(context.Background())
	if err != nil {
		log.Error(errors.Wrap(err, "failed to gracefully stop server"))
	}
	log.Println("server has been stopped")
}

// Start's sever waiting for incoming requests
func (s *Server) Start() {
	log.Printf("starting server on %s", s.Addr)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal("unable to start server: ", err)
			time.Sleep(time.Second)
		}
	}()
}

// NotFound404Handler provides a 404/not found route
func notFound404Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write([]byte(`{"statusCode": 404,"error": "Not Found","message": "Not Found"}`)); err != nil {
		log.Error(errors.Wrap(err, "failed to write response json to output"))
	}
}

// HealthHandler provides a health check route
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
		log.Error(errors.Wrap(err, "failed to write response json to output"))
	}
}
