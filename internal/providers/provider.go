package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"desafio-api/internal/config"
	"desafio-api/internal/domain"
)

type ProviderConfig struct {
	Name                string
	BaseURL             string
	ChargeEndpoint      string
	RefundEndpoint      string
	GetChargeEndpoint   string
	RequestTransformer  func(domain.PaymentRequest) (interface{}, error)
	ResponseTransformer func(*Provider) func([]byte) (*domain.Payment, error)
}

type Provider struct {
	ID         string
	Name       string
	config     ProviderConfig
	httpClient *http.Client
}

func NewProvider(id string, providerConfig ProviderConfig, cfg *config.Config) *Provider {
	return &Provider{
		ID:         id,
		Name:       providerConfig.Name,
		config:     providerConfig,
		httpClient: &http.Client{Timeout: cfg.GetHTTPTimeout()},
	}
}

func (p *Provider) ProcessPayment(request domain.PaymentRequest) (*domain.Payment, error) {
	payload, err := p.config.RequestTransformer(request)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error transforming request: %w", p.Name, err)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error marshaling request: %w", p.Name, err)
	}

	resp, err := p.httpClient.Post(
		fmt.Sprintf("%s%s", p.config.BaseURL, p.config.ChargeEndpoint),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error making request: %w", p.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[provider: %s] unexpected status code: %d", p.Name, resp.StatusCode)
	}

	respBody, err := readBody(resp)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error reading response body: %w", p.Name, err)
	}

	transformer := p.config.ResponseTransformer(p)
	payment, err := transformer(respBody)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error transforming response: %w", p.Name, err)
	}

	return payment, nil
}

func (p *Provider) RefundPayment(paymentID string, request domain.RefundRequest) (*domain.Payment, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error marshaling request: %w", p.Name, err)
	}

	endpoint := strings.ReplaceAll(p.config.RefundEndpoint, "{id}", paymentID)
	resp, err := p.httpClient.Post(
		fmt.Sprintf("%s%s", p.config.BaseURL, endpoint),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error making request: %w", p.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[provider: %s] unexpected status code: %d", p.Name, resp.StatusCode)
	}

	respBody, err := readBody(resp)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error reading response body: %w", p.Name, err)
	}

	transformer := p.config.ResponseTransformer(p)
	payment, err := transformer(respBody)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error transforming response: %w", p.Name, err)
	}

	return payment, nil
}

func (p *Provider) GetPayment(paymentID string) (*domain.Payment, error) {
	endpoint := strings.ReplaceAll(p.config.GetChargeEndpoint, "{id}", paymentID)
	resp, err := p.httpClient.Get(fmt.Sprintf("%s%s", p.config.BaseURL, endpoint))
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error making request: %w", p.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[provider: %s] unexpected status code: %d", p.Name, resp.StatusCode)
	}

	respBody, err := readBody(resp)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error reading response body: %w", p.Name, err)
	}

	transformer := p.config.ResponseTransformer(p)
	payment, err := transformer(respBody)
	if err != nil {
		return nil, fmt.Errorf("[provider: %s] error transforming response: %w", p.Name, err)
	}

	return payment, nil
}

func (p *Provider) GetID() string {
	return p.ID
}

func (p *Provider) GetName() string {
	return p.Name
}

func readBody(resp *http.Response) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
