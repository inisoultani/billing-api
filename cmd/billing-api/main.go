package main

import (
	"billing-api/internal/config"
	billingApiHttp "billing-api/internal/http"
	"billing-api/internal/infra/db"
	"billing-api/internal/infra/db/repository"
	"billing-api/internal/logger"
	"billing-api/internal/service"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	appLogger, logLevel := logger.NewLogger(cfg.AppEnv == "production")
	cfg.LogLevel = logLevel

	appLogger.Info("starting server", slog.String("env", cfg.AppEnv))
	appLogger.Info("Try to establsing db connection...")
	pool, err := db.NewPostgresPool(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to db", slog.Any("err", err))
		os.Exit(1)
	} else {
		appLogger.Info("Successfuly establsing db connection...")
	}

	defer pool.Close()

	billingService := service.NewBillingService(pool, repository.NewPostgresRepo(pool))

	addr := ":" + cfg.ServerPort

	router := billingApiHttp.NewRouter(billingService, cfg)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		appLogger.Info("billing-api started", slog.String("port", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("List error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	appLogger.Info("closing all db connections...")
	pool.Close()

	appLogger.Info("shutting down billing-api server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("List error", slog.Any("err", err))
		os.Exit(1)
	}
	appLogger.Info("server gracefully stopped")
}
