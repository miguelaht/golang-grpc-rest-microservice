package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miguelaht/microservices/order/internal/ports"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *http.Server
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

func (a *Adapter) Run() error {
	router := gin.Default()

	// Register routes
	router.POST("/Order/Create", a.createOrderHandler)

	// Create HTTP server
	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("REST server starting on port %d", a.port)

	// ListenAndServe blocks until server is shut down
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

// Handler methods
func (a *Adapter) createOrderHandler(c *gin.Context) {
	var order CreateOrderRequest
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	res, err := a.Create(c.Request.Context(), &order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
