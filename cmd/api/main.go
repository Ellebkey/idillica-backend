// main.go ≈ src/index.ts: load config, connect databases, build the server
// and listen — plus graceful shutdown, which Node's http.Server hid from you
// and Go writes out explicitly.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"idilica-backend-go/internal/config"
	"idilica-backend-go/internal/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := config.NewLogger(cfg)

	db, err := config.NewDatabase(cfg, logger)
	if err != nil {
		logger.Error("Unable to connect to the database", "error", err)
		os.Exit(1)
	}

	redisClient := config.NewRedis(cfg, logger)

	router := routes.New(cfg, logger, db, redisClient)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		logger.Info(fmt.Sprintf("Listening on port %d", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on Ctrl-C / SIGTERM: stop accepting connections and
	// give in-flight requests up to 10s to finish.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
	_ = redisClient.Close()
	logger.Info("bye")
}
