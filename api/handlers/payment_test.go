package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"desafio-api/internal/domain"
)

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) ProcessPayment(request domain.PaymentRequest) (*domain.Payment, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentService) RefundPayment(paymentID string, request domain.RefundRequest) (*domain.Payment, error) {
	args := m.Called(paymentID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentService) GetPayment(paymentID string) (*domain.Payment, error) {
	args := m.Called(paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func setupRouter(service *MockPaymentService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewPaymentHandler(service)

	router.POST("/payments", handler.ProcessPayment)
	router.POST("/refund/:id", handler.RefundPayment)
	router.GET("/payments/:id", handler.GetPayment)

	return router
}

func TestPaymentHandler_ProcessPayment(t *testing.T) {
	service := new(MockPaymentService)
	router := setupRouter(service)

	t.Run("successful payment", func(t *testing.T) {
		card := domain.Card{
			Number:         gofakeit.CreditCardNumber(nil),
			HolderName:     gofakeit.Name(),
			CVV:            gofakeit.CreditCardCvv(),
			ExpirationDate: gofakeit.CreditCardExp(),
			Installments:   gofakeit.Number(1, 12),
		}

		request := domain.PaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
			Card:        card,
		}

		expectedPayment := &domain.Payment{
			ID:             gofakeit.UUID(),
			CreatedAt:      time.Now(),
			Status:         domain.StatusAuthorized,
			OriginalAmount: request.Amount,
			CurrentAmount:  request.Amount,
			Currency:       request.Currency,
			Description:    request.Description,
			PaymentMethod:  "card",
			CardID:         gofakeit.UUID(),
		}

		service.On("ProcessPayment", request).Return(expectedPayment, nil)

		jsonData, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Payment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedPayment.ID, response.ID)
		assert.Equal(t, expectedPayment.Status, response.Status)
		assert.Equal(t, expectedPayment.OriginalAmount, response.OriginalAmount)

		service.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		invalidJSON := []byte(`{"amount": "invalid"}`)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		request := domain.PaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
		}

		service.On("ProcessPayment", request).Return(nil, errors.New("service error"))

		jsonData, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		service.AssertExpectations(t)
	})
}

func TestPaymentHandler_RefundPayment(t *testing.T) {
	service := new(MockPaymentService)
	router := setupRouter(service)

	t.Run("successful refund", func(t *testing.T) {
		paymentID := gofakeit.UUID()
		request := domain.RefundRequest{
			Amount: gofakeit.Price(10, 500),
		}

		expectedPayment := &domain.Payment{
			ID:             paymentID,
			CreatedAt:      time.Now(),
			Status:         domain.StatusRefunded,
			OriginalAmount: request.Amount * 2,
			CurrentAmount:  request.Amount,
			Currency:       gofakeit.CurrencyShort(),
			Description:    gofakeit.Sentence(4),
			PaymentMethod:  "card",
			CardID:         gofakeit.UUID(),
		}

		service.On("RefundPayment", paymentID, request).Return(expectedPayment, nil)

		jsonData, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refund/"+paymentID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Payment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedPayment.ID, response.ID)
		assert.Equal(t, expectedPayment.Status, response.Status)
		assert.Equal(t, expectedPayment.CurrentAmount, response.CurrentAmount)

		service.AssertExpectations(t)
	})
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	service := new(MockPaymentService)
	router := setupRouter(service)

	t.Run("successful get", func(t *testing.T) {
		paymentID := gofakeit.UUID()
		expectedPayment := &domain.Payment{
			ID:             paymentID,
			CreatedAt:      time.Now(),
			Status:         domain.StatusAuthorized,
			OriginalAmount: gofakeit.Price(10, 1000),
			CurrentAmount:  gofakeit.Price(10, 1000),
			Currency:       gofakeit.CurrencyShort(),
			Description:    gofakeit.Sentence(4),
			PaymentMethod:  "card",
			CardID:         gofakeit.UUID(),
		}

		service.On("GetPayment", paymentID).Return(expectedPayment, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/payments/"+paymentID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response domain.Payment
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedPayment.ID, response.ID)
		assert.Equal(t, expectedPayment.Status, response.Status)
		assert.Equal(t, expectedPayment.OriginalAmount, response.OriginalAmount)

		service.AssertExpectations(t)
	})

	t.Run("payment not found", func(t *testing.T) {
		paymentID := "non-existent"
		service.On("GetPayment", paymentID).Return(nil, errors.New("payment not found"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/payments/"+paymentID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		service.AssertExpectations(t)
	})
}
