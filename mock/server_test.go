package mock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"desafio-api/internal/providers"
)

func TestMockServer(t *testing.T) {
	server := NewMockServer()

	t.Run("process payment successfully", func(t *testing.T) {
		request := providers.MockPaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
			Card: struct {
				Number         string `json:"number"`
				HolderName     string `json:"holderName"`
				CVV            string `json:"cvv"`
				ExpirationDate string `json:"expirationDate"`
				Installments   int    `json:"installments"`
			}{
				Number:         gofakeit.CreditCardNumber(nil),
				HolderName:     gofakeit.Name(),
				CVV:            gofakeit.CreditCardCvv(),
				ExpirationDate: gofakeit.CreditCardExp(),
				Installments:   gofakeit.Number(1, 12),
			},
		}

		jsonData, err := json.Marshal(request)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/charges", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response providers.MockPaymentResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "authorized", response.Status)
		assert.Equal(t, request.Amount, response.OriginalAmount)
		assert.Equal(t, request.Currency, response.Currency)
		assert.Equal(t, request.Description, response.Description)
	})

	t.Run("refund payment successfully", func(t *testing.T) {
		paymentReq := providers.MockPaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
		}

		w := httptest.NewRecorder()
		jsonData, _ := json.Marshal(paymentReq)
		req, _ := http.NewRequest("POST", "/charges", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)

		var payment providers.MockPaymentResponse
		json.Unmarshal(w.Body.Bytes(), &payment)

		refundReq := struct {
			Amount float64 `json:"amount"`
		}{
			Amount: gofakeit.Price(1, paymentReq.Amount-1),
		}

		w = httptest.NewRecorder()
		jsonData, _ = json.Marshal(refundReq)
		req, _ = http.NewRequest("POST", "/refund/"+payment.ID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var refundedPayment providers.MockPaymentResponse
		err := json.Unmarshal(w.Body.Bytes(), &refundedPayment)
		assert.NoError(t, err)
		assert.Equal(t, "refunded", refundedPayment.Status)
		assert.Equal(t, payment.OriginalAmount, refundedPayment.OriginalAmount)
		assert.Equal(t, payment.OriginalAmount-refundReq.Amount, refundedPayment.CurrentAmount)
	})

	t.Run("get payment successfully", func(t *testing.T) {
		paymentReq := providers.MockPaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
		}

		w := httptest.NewRecorder()
		jsonData, _ := json.Marshal(paymentReq)
		req, _ := http.NewRequest("POST", "/charges", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)

		var payment providers.MockPaymentResponse
		json.Unmarshal(w.Body.Bytes(), &payment)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/charges/"+payment.ID, nil)
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedPayment providers.MockPaymentResponse
		err := json.Unmarshal(w.Body.Bytes(), &retrievedPayment)
		assert.NoError(t, err)
		assert.Equal(t, payment.ID, retrievedPayment.ID)
		assert.Equal(t, payment.CurrentAmount, retrievedPayment.CurrentAmount)
	})

	t.Run("simulate failure", func(t *testing.T) {
		server.SimulateFailure(true)
		defer server.SimulateFailure(false)

		request := providers.MockPaymentRequest{
			Amount:      gofakeit.Price(10, 1000),
			Currency:    gofakeit.CurrencyShort(),
			Description: gofakeit.Sentence(3),
		}

		w := httptest.NewRecorder()
		jsonData, _ := json.Marshal(request)
		req, _ := http.NewRequest("POST", "/charges", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
