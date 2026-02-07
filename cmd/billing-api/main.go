package main

import (
	"billing-api/internal/config"
	billingApiHttp "billing-api/internal/http"
	"billing-api/internal/infra/db"
	"billing-api/internal/service"
	"context"
	"log"
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

	log.Printf("Try to establsing db connection...")
	pool, err := db.NewPostgresPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to db : %v", err)
	} else {
		log.Printf("Successfuly establsing db connection...")
	}

	defer pool.Close()

	billingService := service.NewBillingService(pool, db.NewPostgresRepo(pool))

	addr := ":" + cfg.ServerPort

	router := billingApiHttp.NewRouter(billingService, cfg)

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("billing-api listening on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("closing all db connections...")
	pool.Close()

	log.Println("shutting down billing-api server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server gracefully stopped")
}
