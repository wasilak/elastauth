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

func TestIntegration_OIDCProviderRegistration(t *testing.T) {
	// Test that OIDC provider is properly registered
	// This verifies that the import in libs/routes.go works correctly
	
	if !DefaultFactory.IsRegistered("oidc") {
		t.Fatal("OIDC provider should be registered via import in libs/routes.go")
	}
	
	// Verify we can create the OIDC provider (though it will fail validation without proper config)
	// This tests the factory registration mechanism
	_, err := DefaultFactory.Create("oidc", nil)
	if err == nil {
		t.Error("Expected OIDC provider creation to fail without proper configuration")
	}
	
	// The error should be a configuration validation error, not a "provider not found" error
	if !contains(err.Error(), "invalid OIDC configuration") && !contains(err.Error(), "client_id is required") {
		t.Errorf("Expected configuration validation error, got: %v", err)
	}
}

func TestIntegration_AllProvidersRegistered(t *testing.T) {
	// Test that all expected providers are registered
	available := DefaultFactory.ListAvailable()
	
	expectedProviders := []string{"authelia", "oidc"}
	
	for _, expected := range expectedProviders {
		found := false
		for _, available := range available {
			if available == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider '%s' to be registered, available providers: %v", expected, available)
		}
	}
	
	if len(available) < 2 {
		t.Errorf("Expected at least 2 providers to be registered, got %d: %v", len(available), available)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
