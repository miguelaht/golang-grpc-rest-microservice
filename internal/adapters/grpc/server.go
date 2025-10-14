package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/miguelaht/microservices/order/config"
	"github.com/miguelaht/microservices/order/golang/order"
	"github.com/miguelaht/microservices/order/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Adapter struct {
	api        ports.APIPort
	port       int
	grpcServer *grpc.Server
	mu sync.Mutex
	order.UnimplementedOrderServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a *Adapter) Run() error {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}

	var opts []grpc.ServerOption
	opts = append(opts, grpc.Creds(nil))
	a.mu.Lock()
	a.grpcServer = grpc.NewServer(opts...)
	server := a.grpcServer // Keep local reference
	a.mu.Unlock()

	order.RegisterOrderServer(server, a)
	if config.GetEnv() == "development" {
		reflection.Register(server)
	}

	if err := a.grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port")
	}

	return err
}

func (a *Adapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	server := a.grpcServer // Keep local reference
	a.mu.Unlock()

	if server == nil {
		log.Println("Shutting down gRPC server...")
		return nil
	}

	log.Println("Shutting down gRPC server...")

	// Channel to signal when graceful stop is complete
	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or context timeout
	select {
	case <-ctx.Done():
		// Timeout exceeded, force stop
		log.Println("gRPC graceful shutdown timeout, forcing stop...")
		server.Stop()
		return ctx.Err()
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
		return nil
	}
}
