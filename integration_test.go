package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/provider"
)

func TestPhase1Integration_AutheliaProviderWorks(t *testing.T) {
	// Set up Viper configuration to match existing elastauth defaults
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	// Test that Authelia provider is registered (imported via libs package)
	if !provider.DefaultFactory.IsRegistered("authelia") {
		t.Fatal("Authelia provider should be registered when libs package is imported")
	}
	
	// Create provider using factory (same as getAuthProvider in routes.go)
	authProvider, err := provider.DefaultFactory.Create("authelia", nil)
	if err != nil {
		t.Fatalf("Failed to create Authelia provider: %v", err)
	}
	
	if authProvider.Type() != "authelia" {
		t.Errorf("Expected provider type 'authelia', got '%s'", authProvider.Type())
	}
	
	// Test with typical Authelia headers
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Remote-User", "john.doe")
	req.Header.Set("Remote-Groups", "admin,users")
	req.Header.Set("Remote-Email", "john.doe@example.com")
	req.Header.Set("Remote-Name", "John Doe")
	
	authReq := &provider.AuthRequest{Request: req}
	userInfo, err := authProvider.GetUser(context.Background(), authReq)
	if err != nil {
		t.Fatalf("Failed to get user info: %v", err)
	}
	
	// Verify all fields are extracted correctly (backward compatibility)
	if userInfo.Username != "john.doe" {
		t.Errorf("Expected username 'john.doe', got '%s'", userInfo.Username)
	}
	
	if userInfo.Email != "john.doe@example.com" {
		t.Errorf("Expected email 'john.doe@example.com', got '%s'", userInfo.Email)
	}
	
	if userInfo.FullName != "John Doe" {
		t.Errorf("Expected full name 'John Doe', got '%s'", userInfo.FullName)
	}
	
	expectedGroups := []string{"admin", "users"}
	if len(userInfo.Groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(userInfo.Groups))
	}
	
	for i, expected := range expectedGroups {
		if i >= len(userInfo.Groups) || userInfo.Groups[i] != expected {
			t.Errorf("Expected group[%d] = '%s', got '%s'", i, expected, userInfo.Groups[i])
		}
	}
}

func TestPhase1Integration_BackwardCompatibilityValidation(t *testing.T) {
	// Test that existing validation functions still work
	// This ensures the provider system doesn't break existing validation
	
	// Test username validation
	if err := libs.ValidateUsername("john.doe"); err != nil {
		t.Errorf("Valid username should pass validation: %v", err)
	}
	
	if err := libs.ValidateUsername("invalid user!"); err == nil {
		t.Error("Invalid username should fail validation")
	}
	
	// Test email validation
	if err := libs.ValidateEmail("john.doe@example.com"); err != nil {
		t.Errorf("Valid email should pass validation: %v", err)
	}
	
	if err := libs.ValidateEmail("invalid-email"); err == nil {
		t.Error("Invalid email should fail validation")
	}
	
	// Test name validation
	if err := libs.ValidateName("John Doe"); err != nil {
		t.Errorf("Valid name should pass validation: %v", err)
	}
	
	// Test group parsing and validation
	groups, err := libs.ParseAndValidateGroups("admin,users", false, nil)
	if err != nil {
		t.Errorf("Valid groups should parse successfully: %v", err)
	}
	
	expectedGroups := []string{"admin", "users"}
	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}
}