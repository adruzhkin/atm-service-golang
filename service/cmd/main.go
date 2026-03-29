package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adruzhkin/atm-service-golang/service/db"
	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/server"
)

var (
	jwtSecret        = os.Getenv("JWT_SECRET")
	jwtExpiry        = os.Getenv("JWT_EXPIRY")
	jwtRefreshExpiry = os.Getenv("JWT_REFRESH_EXPIRY")
	port             = flag.Int("port", 8080, "http port to listen on")
	timeout          = flag.Duration("timeout", 5*time.Second, "timeout for graceful shutdown")
)

func main() {
	flag.Parse()

	accessExpiry := parseDurationOrDefault(jwtExpiry, 15*time.Minute)
	refreshExpiry := parseDurationOrDefault(jwtRefreshExpiry, 168*time.Hour)

	s := server.New(port)
	s.JWT = &jwt.Token{
		Secret:             jwtSecret,
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
	}
	s.DB = &db.Postgres{}

	for i := 0; i < 3; i++ {
		// Pause before connection to wait for db docker container build
		time.Sleep(5 * time.Second)
		err := s.DB.Open()
		if err != nil {
			slog.Error("cannot open db connection", "error", err)
			continue
		}

		break
	}
	defer s.DB.Close()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server is up and running", "port", *port)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("unexpected shutdown", "error", err)
			os.Exit(1)
		}
	}()

	<-stopCh
	s.ShutdownGracefully(timeout)
}

func parseDurationOrDefault(val string, fallback time.Duration) time.Duration {
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		slog.Warn("invalid duration, using default", "value", val, "default", fallback)
		return fallback
	}
	return d
}
