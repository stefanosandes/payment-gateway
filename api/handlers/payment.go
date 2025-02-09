package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"desafio-api/internal/domain"
)

type PaymentService interface {
	ProcessPayment(request domain.PaymentRequest) (*domain.Payment, error)
	RefundPayment(paymentID string, request domain.RefundRequest) (*domain.Payment, error)
	GetPayment(paymentID string) (*domain.Payment, error)
}

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var request domain.PaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	payment, err := h.service.ProcessPayment(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	var request domain.RefundRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	payment, err := h.service.RefundPayment(paymentID, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refund payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	payment, err := h.service.GetPayment(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}
