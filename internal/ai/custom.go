package ai

import (
	"context"
	"fmt"

	"github.com/paradoxe35/encre/internal/config"
)

type CustomProvider struct {
	name         string
	providerType string
	inner        Provider
}

func NewCustomProvider(name, providerType, apiKey, baseURL, model string, temperature float64) (*CustomProvider, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required for custom providers")
	}
	// Custom providers always use OpenAI-compatible API
	providerType = config.ProviderTypeOpenAICompatible

	return &CustomProvider{
		name:         name,
		providerType: providerType,
		inner:        NewOpenAIProvider(apiKey, baseURL, model, temperature),
	}, nil
}

func (p *CustomProvider) ReviseText(ctx context.Context, text, systemPrompt string) (string, error) {
	return p.inner.ReviseText(ctx, text, systemPrompt)
}

func (p *CustomProvider) SetLowReasoning(low bool) {
	if aware, ok := p.inner.(ReasoningAware); ok {
		aware.SetLowReasoning(low)
	}
}

func (p *CustomProvider) ValidateConfig() error {
	return p.inner.ValidateConfig()
}

func (p *CustomProvider) GetName() string {
	return p.name
}

func (p *CustomProvider) GetModel() string {
	return p.inner.GetModel()
}

func (p *CustomProvider) GetTemperature() float64 {
	return p.inner.GetTemperature()
}

func (p *CustomProvider) GetProviderType() string {
	return p.providerType
}
