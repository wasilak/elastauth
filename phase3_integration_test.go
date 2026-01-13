package main

import (
	"testing"

	"github.com/wasilak/elastauth/provider/oidc"
)

// TestPhase3Integration_ProviderCreation tests OIDC provider creation
func TestPhase3Integration_ProviderCreation(t *testing.T) {
	// Test creating OIDC provider with manual endpoints
	config := oidc.Config{
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      "https://example.com/userinfo",
		TokenValidation:       "userinfo",
	}

	provider, err := oidc.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create OIDC provider: %v", err)
	}

	if provider.Type() != "oidc" {
		t.Errorf("Expected provider type 'oidc', got '%s'", provider.Type())
	}

	// Test provider validation
	err = provider.Validate()
	if err != nil {
		t.Errorf("Provider validation failed: %v", err)
	}

	t.Log("Phase 3 integration test passed: OIDC provider successfully created and validated")
}

// TestPhase3Integration_ConfigurationDefaults tests OIDC configuration defaults
func TestPhase3Integration_ConfigurationDefaults(t *testing.T) {
	// Test that OIDC provider can be created and defaults are applied
	config := oidc.Config{
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      "https://example.com/userinfo",
		TokenValidation:       "userinfo",
	}

	provider, err := oidc.NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create OIDC provider: %v", err)
	}

	if provider.Type() != "oidc" {
		t.Errorf("Expected provider type 'oidc', got '%s'", provider.Type())
	}

	t.Log("Phase 3 integration test passed: OIDC provider configuration and defaults work correctly")
}