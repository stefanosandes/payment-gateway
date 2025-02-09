package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"desafio-api/api/handlers"
	"desafio-api/internal/config"
	"desafio-api/internal/domain"
	"desafio-api/internal/providers"
	"desafio-api/internal/service"
	"desafio-api/mock"
)

func main() {

	// Load the configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start server to simulate the two payment providers
	mockServer1 := mock.NewMockServer()
	go func() {
		if err := mockServer1.Run(":3001"); err != nil {
			log.Fatalf("Failed to start mock server 1: %v", err)
		}
	}()

	mockServer2 := mock.NewMockServer()
	go func() {
		if err := mockServer2.Run(":3002"); err != nil {
			log.Fatalf("Failed to start mock server 2: %v", err)
		}
	}()

	// Configure payment providers
	provider1 := providers.NewProvider("stripe", providers.ProviderConfig{
		Name:                "Stripe",
		BaseURL:             "http://localhost:3001",
		ChargeEndpoint:      "/charges",
		RefundEndpoint:      "/refund/{id}",
		GetChargeEndpoint:   "/charges/{id}",
		RequestTransformer:  providers.StandardRequestTransformer,
		ResponseTransformer: providers.StandardResponseTransformer,
	}, cfg)

	provider2 := providers.NewProvider("braintree", providers.ProviderConfig{
		Name:                "Braintree",
		BaseURL:             "http://localhost:3002",
		ChargeEndpoint:      "/charges",
		RefundEndpoint:      "/refund/{id}",
		GetChargeEndpoint:   "/charges/{id}",
		RequestTransformer:  providers.StandardRequestTransformer,
		ResponseTransformer: providers.StandardResponseTransformer,
	}, cfg)

	// Create payment service and handler
	var paymentService handlers.PaymentService = service.NewPaymentService([]domain.PaymentProvider{provider1, provider2}, cfg)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// Setup routes
	router := gin.Default()
	router.POST("/payments", paymentHandler.ProcessPayment)
	router.POST("/refund/:id", paymentHandler.RefundPayment)
	router.GET("/payments/:id", paymentHandler.GetPayment)

	// Start the server
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to start the API server: %v", err)
		}
	}()

	log.Println("API is running on port :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down all servers...")
}
