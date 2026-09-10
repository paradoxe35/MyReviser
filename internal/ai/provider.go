package ai

import (
	"context"
	"fmt"
	"sync"
)

// Provider defines the interface for AI text revision providers
type Provider interface {
	ReviseText(ctx context.Context, text, systemPrompt string) (string, error)
	ValidateConfig() error
	GetName() string
	GetModel() string
	GetTemperature() float64
}

// ProviderFactory manages AI providers
type ProviderFactory struct {
	mu        sync.RWMutex
	providers map[string]Provider
	current   Provider
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]Provider),
	}
}

// Register adds a new provider to the factory
func (f *ProviderFactory) Register(name string, provider Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = provider
}

// Reset drops every cached provider. Settings changes invalidate them all:
// a stale entry would keep using the previous API key or model.
func (f *ProviderFactory) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers = make(map[string]Provider)
	f.current = nil
}

// Get returns a provider by name
func (f *ProviderFactory) Get(name string) (Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// SetCurrent sets the current active provider
func (f *ProviderFactory) SetCurrent(name string) error {
	provider, err := f.Get(name)
	if err != nil {
		return err
	}
	if err := provider.ValidateConfig(); err != nil {
		return fmt.Errorf("provider validation failed: %w", err)
	}
	f.mu.Lock()
	f.current = provider
	f.mu.Unlock()
	return nil
}

// GetCurrent returns the current active provider
func (f *ProviderFactory) GetCurrent() Provider {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.current
}
