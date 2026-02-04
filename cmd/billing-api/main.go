package main

import (
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

	dsn := "postgres://billing:billing@localhost:5432/billing?sslmode=disable"

	pool, err := db.NewPostgresPool(dsn)
	if err != nil {
		log.Fatal("Failed to connect to db : %v", err)
	}

	defer pool.Close()

	billingService := service.NewBillingService(pool)

	addr := ":8081"

	router := billingApiHttp.NewRouter(billingService)

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

	log.Println("shutting down billing-api server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	log.Println("server gracefully stopped")
}
