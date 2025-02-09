package service

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"desafio-api/internal/config"
	"desafio-api/internal/domain"
)

func getTestConfig() *config.Config {
	return &config.Config{
		HTTP: config.HTTPConfig{
			TimeoutSeconds: 10,
		},
		Retry: config.RetryConfig{
			Attempts:     3,
			DelaySeconds: 1,
		},
		CircuitBreaker: config.CircuitBreakerConfig{
			MaxRequests:     3,
			IntervalSeconds: 10,
			TimeoutSeconds:  30,
			MinRequests:     3,
			FailureRatio:    0.6,
		},
	}
}

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) ProcessPayment(request domain.PaymentRequest) (*domain.Payment, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockProvider) RefundPayment(paymentID string, request domain.RefundRequest) (*domain.Payment, error) {
	args := m.Called(paymentID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockProvider) GetPayment(paymentID string) (*domain.Payment, error) {
	args := m.Called(paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockProvider) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func TestPaymentServiceRefund(t *testing.T) {
	gofakeit.Seed(0)

	t.Run("successful refund", func(t *testing.T) {
		provider1 := new(MockProvider)
		provider1.On("GetID").Return("stripe").Maybe()
		provider1.On("GetName").Return("Stripe").Maybe()

		provider2 := new(MockProvider)
		provider2.On("GetID").Return("braintree").Maybe()
		provider2.On("GetName").Return("Braintree").Maybe()

		cardID := gofakeit.CreditCardNumber(nil)

		originalPayment := &domain.Payment{
			ID:             gofakeit.UUID(),
			CreatedAt:      time.Now(),
			Status:         domain.StatusAuthorized,
			OriginalAmount: gofakeit.Price(50, 200),
			CurrentAmount:  gofakeit.Price(50, 200),
			Currency:       gofakeit.CurrencyShort(),
			Description:    gofakeit.Sentence(3),
			PaymentMethod:  "card",
			CardID:         cardID,
		}

		refundRequest := domain.RefundRequest{
			Amount: originalPayment.CurrentAmount,
		}

		refundedPayment := &domain.Payment{
			ID:             originalPayment.ID,
			CreatedAt:      time.Now(),
			Status:         domain.StatusRefunded,
			OriginalAmount: originalPayment.OriginalAmount,
			CurrentAmount:  0.0,
			Currency:       originalPayment.Currency,
			Description:    originalPayment.Description,
			PaymentMethod:  originalPayment.PaymentMethod,
			CardID:         cardID,
		}

		provider1.On("RefundPayment", originalPayment.ID, refundRequest).Return(refundedPayment, nil)

		service := NewPaymentService([]domain.PaymentProvider{provider1, provider2}, getTestConfig())
		service.transactions[originalPayment.ID] = &domain.Transaction{
			Payment:      originalPayment,
			ProviderID:   "stripe",
			ProviderName: "Stripe",
		}

		payment, err := service.RefundPayment(originalPayment.ID, refundRequest)

		assert.NoError(t, err)
		assert.Equal(t, domain.StatusRefunded, payment.Status)
		assert.Equal(t, float64(0), payment.CurrentAmount)
		mock.AssertExpectationsForObjects(t, provider1, provider2)
	})

	t.Run("refund non-existent payment", func(t *testing.T) {
		provider := new(MockProvider)
		service := NewPaymentService([]domain.PaymentProvider{provider}, getTestConfig())

		refundRequest := domain.RefundRequest{
			Amount: gofakeit.Price(50, 200),
		}

		payment, err := service.RefundPayment("non-existent", refundRequest)

		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "payment not found")
		provider.AssertNotCalled(t, "RefundPayment")
	})

	t.Run("refund non-authorized payment", func(t *testing.T) {
		provider := new(MockProvider)
		provider.On("GetID").Return("stripe")
		provider.On("GetName").Return("Stripe")

		cardID := gofakeit.CreditCardNumber(nil)
		failedPayment := &domain.Payment{
			ID:             gofakeit.UUID(),
			CreatedAt:      time.Now(),
			Status:         domain.StatusFailed,
			OriginalAmount: gofakeit.Price(50, 200),
			CurrentAmount:  gofakeit.Price(50, 200),
			Currency:       gofakeit.CurrencyShort(),
			Description:    gofakeit.LoremIpsumSentence(5),
			PaymentMethod:  "card",
			CardID:         cardID,
		}

		service := NewPaymentService([]domain.PaymentProvider{provider}, getTestConfig())
		service.transactions[failedPayment.ID] = &domain.Transaction{
			Payment:      failedPayment,
			ProviderID:   "stripe",
			ProviderName: "Stripe",
		}

		refundRequest := domain.RefundRequest{
			Amount: gofakeit.Price(20, 50),
		}

		payment, err := service.RefundPayment(failedPayment.ID, refundRequest)

		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "payment cannot be refunded")
		provider.AssertNotCalled(t, "RefundPayment")
	})
}
