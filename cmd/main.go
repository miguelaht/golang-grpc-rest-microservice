package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/miguelaht/microservices/order/config"
	"github.com/miguelaht/microservices/order/internal/adapters/db"
	"github.com/miguelaht/microservices/order/internal/adapters/grpc"
	"github.com/miguelaht/microservices/order/internal/adapters/payment"
	"github.com/miguelaht/microservices/order/internal/adapters/rest"
	"github.com/miguelaht/microservices/order/internal/application/core/api"
	"golang.org/x/sync/errgroup"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}

	paymentAdapter, err := payment.NewAdapter(config.GetPaymentServiceUrl())
	if err != nil {
		log.Fatalf("Failed to initialize payment stub. Error: %v", err)
	}

	application := api.NewApplication(dbAdapter, paymentAdapter)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	g, gCtx := errgroup.WithContext(ctx)

	grpcAdapter := grpc.NewAdapter(application, config.GetGRPCPort())
	g.Go(func() error {
		log.Println("Starting gRPC server...")
		return grpcAdapter.Run()
	})

	restAdapter := rest.NewAdapter(application, config.GetHTTPPort())
	g.Go(func() error {
		log.Println("Starting REST server...")
		return restAdapter.Run()
	})

	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Shutting down...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cancel()

		grpcAdapter.Shutdown(shutdownCtx)
		restAdapter.Shutdown(shutdownCtx)

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
