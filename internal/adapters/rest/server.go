package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/miguelaht/microservices/order/golang/order"
	"github.com/miguelaht/microservices/order/internal/ports"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *http.Server
	order.UnimplementedOrderServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

func (a *Adapter) Run() error {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	order.RegisterOrderHandlerServer(ctx, mux, a)

	// Create HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("REST server starting on port %d", a.port)

	// ListenAndServe blocks until server is shut down
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve REST on port %d: %w", a.port, err)
	}
	return nil

}

// Shutdown gracefully stops the REST server
func (a *Adapter) Shutdown(ctx context.Context) error {
	if a.server == nil {
		return nil
	}

	log.Println("Shutting down REST server...")

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("REST server shutdown error: %w", err)
	}

	log.Println("REST server stopped gracefully")
	return nil
}
