package provider

import (
	"context"
	"testing"
)

// mockProvider is a test implementation of AuthProvider
type mockProvider struct {
	providerType string
}

func (m *mockProvider) GetUser(ctx context.Context, req *AuthRequest) (*UserInfo, error) {
	return &UserInfo{
		Username: "testuser",
		Email:    "test@example.com",
		Groups:   []string{"testgroup"},
		FullName: "Test User",
	}, nil
}

func (m *mockProvider) Type() string {
	return m.providerType
}

func (m *mockProvider) Validate() error {
	return nil
}

func mockConstructor(config interface{}) (AuthProvider, error) {
	return &mockProvider{providerType: "mock"}, nil
}

func TestFactory_Register(t *testing.T) {
	factory := NewFactory()
	
	factory.Register("mock", mockConstructor)
	
	if !factory.IsRegistered("mock") {
		t.Error("Expected mock provider to be registered")
	}
}

func TestFactory_Create(t *testing.T) {
	factory := NewFactory()
	factory.Register("mock", mockConstructor)
	
	provider, err := factory.Create("mock", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if provider.Type() != "mock" {
		t.Errorf("Expected provider type 'mock', got '%s'", provider.Type())
	}
}

func TestFactory_Create_UnknownProvider(t *testing.T) {
	factory := NewFactory()
	
	_, err := factory.Create("unknown", nil)
	if err == nil {
		t.Error("Expected error for unknown provider type")
	}
}

func TestFactory_ListAvailable(t *testing.T) {
	factory := NewFactory()
	factory.Register("mock1", mockConstructor)
	factory.Register("mock2", mockConstructor)
	
	available := factory.ListAvailable()
	if len(available) != 2 {
		t.Errorf("Expected 2 available providers, got %d", len(available))
	}
}

func TestDefaultFactory_AutheliaRegistered(t *testing.T) {
	// Test that Authelia provider can be registered
	// Note: The actual registration happens when the authelia package is imported
	// This test verifies the registration mechanism works
	
	// Manually register for this test since we can't import authelia package here
	// due to import cycles
	DefaultFactory.Register("test-authelia", mockConstructor)
	
	if !DefaultFactory.IsRegistered("test-authelia") {
		t.Error("Expected test-authelia provider to be registered in DefaultFactory")
	}
	
	available := DefaultFactory.ListAvailable()
	found := false
	for _, providerType := range available {
		if providerType == "test-authelia" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected test-authelia provider to be in available providers list")
	}
}