package domain

import "time"

type PaymentStatus string

const (
	StatusAuthorized PaymentStatus = "authorized"
	StatusFailed     PaymentStatus = "failed"
	StatusRefunded   PaymentStatus = "refunded"
)

type Card struct {
	Number         string `json:"number"`
	HolderName     string `json:"holderName"`
	CVV            string `json:"cvv"`
	ExpirationDate string `json:"expirationDate"`
	Installments   int    `json:"installments"`
}

type PaymentRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	Card        Card    `json:"card"`
}

type Payment struct {
	ID             string        `json:"id"`
	CreatedAt      time.Time     `json:"createdAt"`
	Status         PaymentStatus `json:"status"`
	OriginalAmount float64       `json:"originalAmount"`
	CurrentAmount  float64       `json:"currentAmount"`
	Currency       string        `json:"currency"`
	Description    string        `json:"description"`
	PaymentMethod  string        `json:"paymentMethod"`
	CardID         string        `json:"cardId"`
}

type Transaction struct {
	Payment      *Payment `json:"payment"`
	ProviderID   string   `json:"providerId"`
	ProviderName string   `json:"providerName"`
}

type RefundRequest struct {
	Amount float64 `json:"amount"`
}

type PaymentProvider interface {
	ProcessPayment(request PaymentRequest) (*Payment, error)
	RefundPayment(paymentID string, request RefundRequest) (*Payment, error)
	GetPayment(paymentID string) (*Payment, error)
	GetID() string
	GetName() string
}
