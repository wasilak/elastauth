package provider

import (
	"fmt"
	"sync"
)

// ProviderConstructor is a function that creates a new provider instance
type ProviderConstructor func(config interface{}) (AuthProvider, error)

// Factory manages provider registration and instantiation
type Factory struct {
	mu        sync.RWMutex
	providers map[string]ProviderConstructor
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		providers: make(map[string]ProviderConstructor),
	}
}

// Register registers a provider constructor with the factory
func (f *Factory) Register(providerType string, constructor ProviderConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[providerType] = constructor
}

// Create creates a new provider instance of the specified type
func (f *Factory) Create(providerType string, config interface{}) (AuthProvider, error) {
	f.mu.RLock()
	constructor, exists := f.providers[providerType]
	f.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
	
	return constructor(config)
}

// ListAvailable returns a list of all registered provider types
func (f *Factory) ListAvailable() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	types := make([]string, 0, len(f.providers))
	for providerType := range f.providers {
		types = append(types, providerType)
	}
	return types
}

// IsRegistered checks if a provider type is registered
func (f *Factory) IsRegistered(providerType string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, exists := f.providers[providerType]
	return exists
}

// DefaultFactory is the global provider factory instance
var DefaultFactory = NewFactory()