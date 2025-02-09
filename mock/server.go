package mock

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"desafio-api/internal/providers"
)

type MockServer struct {
	router      *gin.Engine
	payments    map[string]providers.MockPaymentResponse
	mutex       sync.RWMutex
	failureMode bool
}

func NewMockServer() *MockServer {
	server := &MockServer{
		router:      gin.Default(),
		payments:    make(map[string]providers.MockPaymentResponse),
		failureMode: false,
	}
	server.setupRoutes()
	return server
}

func (s *MockServer) setupRoutes() {
	s.router.POST("/charges", s.handleCharge)
	s.router.POST("/refund/:id", s.handleRefund)
	s.router.GET("/charges/:id", s.handleGetCharge)
}

func (s *MockServer) Run(addr string) error {
	log.Printf("Mock server running on %s", addr)
	return s.router.Run(addr)
}

func (s *MockServer) handleCharge(c *gin.Context) {
	s.mutex.RLock()
	if s.failureMode {
		s.mutex.RUnlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}
	s.mutex.RUnlock()

	var req providers.MockPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simulate processing delay
	time.Sleep(100 * time.Millisecond)

	// Create mock response
	resp := providers.MockPaymentResponse{
		ID:             uuid.New().String(),
		CreatedAt:      time.Now(),
		Status:         "authorized",
		OriginalAmount: req.Amount,
		CurrentAmount:  req.Amount,
		Currency:       req.Currency,
		Description:    req.Description,
		PaymentMethod:  "card",
		CardID:         uuid.New().String(),
	}

	// Store payment
	s.mutex.Lock()
	s.payments[resp.ID] = resp
	s.mutex.Unlock()

	c.JSON(http.StatusOK, resp)
}

func (s *MockServer) handleRefund(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mutex.Lock()
	payment, exists := s.payments[id]
	if !exists {
		s.mutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	// Update payment status and amount
	payment.Status = "refunded"
	payment.CurrentAmount -= req.Amount
	s.payments[id] = payment
	s.mutex.Unlock()

	c.JSON(http.StatusOK, payment)
}

func (s *MockServer) handleGetCharge(c *gin.Context) {
	id := c.Param("id")

	s.mutex.RLock()
	payment, exists := s.payments[id]
	s.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// SimulateFailure allows controlling the mock server's behavior for testing
func (s *MockServer) SimulateFailure(enabled bool) {
	s.mutex.Lock()
	s.failureMode = enabled
	s.mutex.Unlock()
}
