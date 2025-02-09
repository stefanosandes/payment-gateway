package service

import (
	"fmt"
	"log"

	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"

	"desafio-api/internal/config"
	"desafio-api/internal/domain"
)

type PaymentService struct {
	providers      []domain.PaymentProvider
	circuitBreaker *gobreaker.CircuitBreaker
	transactions   map[string]*domain.Transaction
	config         *config.Config
}

func NewPaymentService(providers []domain.PaymentProvider, cfg *config.Config) *PaymentService {
	if len(providers) == 0 {
		panic("At least one payment provider is required")
	}

	settings := gobreaker.Settings{
		Name:        "payment-provider",
		MaxRequests: cfg.CircuitBreaker.MaxRequests,
		Interval:    cfg.GetCircuitBreakerInterval(),
		Timeout:     cfg.GetCircuitBreakerTimeout(),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= cfg.CircuitBreaker.MinRequests && failureRatio >= cfg.CircuitBreaker.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit breaker state changed from %v to %v", from, to)
		},
	}

	return &PaymentService{
		providers:      providers,
		circuitBreaker: gobreaker.NewCircuitBreaker(settings),
		transactions:   make(map[string]*domain.Transaction),
		config:         cfg,
	}
}

func (s *PaymentService) ProcessPayment(request domain.PaymentRequest) (*domain.Payment, error) {
	var lastErr error
	for _, provider := range s.providers {
		log.Printf("[provider: %s] attempting to process payment", provider.GetName())

		result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
			var payment *domain.Payment
			err := retry.Do(
				func() error {
					var err error
					payment, err = provider.ProcessPayment(request)
					if err != nil {
						log.Printf("[provider: %s] attempt failed: %v", provider.GetName(), err)
						return err
					}
					return nil
				},
				retry.Attempts(uint(s.config.Retry.Attempts)),
				retry.Delay(s.config.GetRetryDelay()),
				retry.OnRetry(func(n uint, err error) {
					log.Printf("[provider: %s] retry %d: %v", provider.GetName(), n+1, err)
				}),
			)
			if err != nil {
				return nil, fmt.Errorf("[provider: %s] failed after retries: %w", provider.GetName(), err)
			}
			return payment, nil
		})

		payment, ok := result.(*domain.Payment)

		if err != nil {
			log.Printf("[provider: %s] failed: %v", provider.GetName(), err)
			if ok && payment != nil {
				payment.Status = domain.StatusFailed
				failedProviderID := provider.GetID()
				transaction := &domain.Transaction{
					Payment:      payment,
					ProviderID:   failedProviderID,
					ProviderName: provider.GetName(),
				}
				s.transactions[payment.ID] = transaction
			}
			lastErr = err
			continue
		}

		log.Printf("[provider: %s] payment successfully processed", provider.GetName())
		successProviderID := provider.GetID()
		transaction := &domain.Transaction{
			Payment:      payment,
			ProviderID:   successProviderID,
			ProviderName: provider.GetName(),
		}
		s.transactions[payment.ID] = transaction
		return payment, nil
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

func (s *PaymentService) RefundPayment(paymentID string, request domain.RefundRequest) (*domain.Payment, error) {
	transaction, exists := s.transactions[paymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found: %s", paymentID)
	}

	if transaction.Payment.Status != domain.StatusAuthorized {
		return nil, fmt.Errorf("payment cannot be refunded: status is %s", transaction.Payment.Status)
	}

	var provider domain.PaymentProvider
	for _, p := range s.providers {
		if p.GetID() == transaction.ProviderID {
			provider = p
			break
		}
	}

	if provider == nil {
		return nil, fmt.Errorf("provider not found: %s", transaction.ProviderID)
	}

	log.Printf("[provider: %s] attempting to refund payment", provider.GetName())

	result, err := s.circuitBreaker.Execute(func() (interface{}, error) {
		var payment *domain.Payment
		err := retry.Do(
			func() error {
				var err error
				payment, err = provider.RefundPayment(paymentID, request)
				if err != nil {
					log.Printf("[provider: %s] attempt failed: %v", provider.GetName(), err)
					return err
				}
				return nil
			},
			retry.Attempts(uint(s.config.Retry.Attempts)),
			retry.Delay(s.config.GetRetryDelay()),
			retry.OnRetry(func(n uint, err error) {
				log.Printf("[provider: %s] retry %d: %v", provider.GetName(), n+1, err)
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("[provider: %s] failed after retries: %w", provider.GetName(), err)
		}
		return payment, nil
	})

	payment, ok := result.(*domain.Payment)
	if err != nil {
		log.Printf("[provider: %s] failed: %v", provider.GetName(), err)

		if ok && payment != nil {
			payment.Status = domain.StatusFailed
			transaction.Payment = payment
		}
		return nil, err
	}

	log.Printf("[provider: %s] refund successfully processed", provider.GetName())
	payment.Status = domain.StatusRefunded
	transaction.Payment = payment
	return payment, nil
}

func (s *PaymentService) GetPayment(paymentID string) (*domain.Payment, error) {
	transaction, exists := s.transactions[paymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found: %s", paymentID)
	}
	return transaction.Payment, nil
}
