package providers

import (
	"encoding/json"
	"fmt"
	"time"

	"desafio-api/internal/domain"
)

type MockPaymentRequest struct {
	Amount      float64     `json:"amount"`
	Currency    string      `json:"currency"`
	Description string      `json:"description"`
	Card        domain.Card `json:"card"`
}

type MockPaymentResponse struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	Status         string    `json:"status"`
	OriginalAmount float64   `json:"originalAmount"`
	CurrentAmount  float64   `json:"currentAmount"`
	Currency       string    `json:"currency"`
	Description    string    `json:"description"`
	PaymentMethod  string    `json:"paymentMethod"`
	CardID         string    `json:"cardId"`
}

func StandardRequestTransformer(request domain.PaymentRequest) (interface{}, error) {
	return MockPaymentRequest{
		Amount:      request.Amount,
		Currency:    request.Currency,
		Description: request.Description,
		Card:        request.Card,
	}, nil
}

func StandardResponseTransformer(provider *Provider) func([]byte) (*domain.Payment, error) {
	return func(data []byte) (*domain.Payment, error) {
		var resp MockPaymentResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}

		var status domain.PaymentStatus
		switch resp.Status {
		case "authorized", "paid":
			status = domain.StatusAuthorized
		case "refunded", "voided":
			status = domain.StatusRefunded
		default:
			status = domain.StatusFailed
		}

		payment := &domain.Payment{
			ID:             resp.ID,
			CreatedAt:      resp.CreatedAt,
			Status:         status,
			OriginalAmount: resp.OriginalAmount,
			CurrentAmount:  resp.CurrentAmount,
			Currency:       resp.Currency,
			Description:    resp.Description,
			PaymentMethod:  resp.PaymentMethod,
			CardID:         resp.CardID,
		}
		return payment, nil
	}
}
