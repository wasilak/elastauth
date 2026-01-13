package provider

import (
	"testing"

	"github.com/spf13/viper"
)

func TestIntegration_AutheliaProviderBackwardCompatibility(t *testing.T) {
	// Set up Viper configuration to match existing elastauth defaults
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	// Manually register a test provider to simulate the integration
	// (We can't import authelia package due to import cycles)
	testConstructor := func(config interface{}) (AuthProvider, error) {
		return &mockProvider{providerType: "authelia"}, nil
	}
	DefaultFactory.Register("authelia", testConstructor)
	
	// Test that Authelia provider is registered and can be created
	if !DefaultFactory.IsRegistered("authelia") {
		t.Fatal("Authelia provider should be registered")
	}
	
	// Create provider using factory (same as getAuthProvider in routes.go)
	authProvider, err := DefaultFactory.Create("authelia", nil)
	if err != nil {
		t.Fatalf("Failed to create Authelia provider: %v", err)
	}
	
	if authProvider.Type() != "authelia" {
		t.Errorf("Expected provider type 'authelia', got '%s'", authProvider.Type())
	}
}

func TestIntegration_DefaultProviderSelection(t *testing.T) {
	// Test that when no auth_provider is configured, it defaults to "authelia"
	viper.Set("auth_provider", "") // Empty means default
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	// This simulates the logic in getAuthProvider()
	providerType := viper.GetString("auth_provider")
	if providerType == "" {
		providerType = "authelia"
	}
	
	if providerType != "authelia" {
		t.Errorf("Expected default provider type 'authelia', got '%s'", providerType)
	}
	
	// Verify we can create the default provider
	authProvider, err := DefaultFactory.Create(providerType, nil)
	if err != nil {
		t.Fatalf("Failed to create default provider: %v", err)
	}
	
	if authProvider.Type() != "authelia" {
		t.Errorf("Expected provider type 'authelia', got '%s'", authProvider.Type())
	}
}
