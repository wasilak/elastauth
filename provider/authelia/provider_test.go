package authelia

import (
	"context"
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/provider"
)

func TestProvider_GetUser(t *testing.T) {
	// Set up Viper configuration for testing
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	p, err := NewProvider(nil)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a mock HTTP request with Authelia headers
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "group1,group2")
	req.Header.Set("Remote-Email", "test@example.com")
	req.Header.Set("Remote-Name", "Test User")

	authReq := &provider.AuthRequest{Request: req}
	
	userInfo, err := p.GetUser(context.Background(), authReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if userInfo.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", userInfo.Username)
	}

	if userInfo.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", userInfo.Email)
	}

	if userInfo.FullName != "Test User" {
		t.Errorf("Expected full name 'Test User', got '%s'", userInfo.FullName)
	}

	if len(userInfo.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(userInfo.Groups))
	}

	if userInfo.Groups[0] != "group1" || userInfo.Groups[1] != "group2" {
		t.Errorf("Expected groups [group1, group2], got %v", userInfo.Groups)
	}
}

func TestProvider_GetUser_MissingUsername(t *testing.T) {
	// Set up Viper configuration for testing
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	p, err := NewProvider(nil)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a mock HTTP request without username header
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Remote-Groups", "group1")
	req.Header.Set("Remote-Email", "test@example.com")

	authReq := &provider.AuthRequest{Request: req}
	
	_, err = p.GetUser(context.Background(), authReq)
	if err == nil {
		t.Error("Expected error for missing username header")
	}
}

func TestProvider_Type(t *testing.T) {
	// Set up Viper configuration for testing
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	
	p, err := NewProvider(nil)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if p.Type() != "authelia" {
		t.Errorf("Expected provider type 'authelia', got '%s'", p.Type())
	}
}

func TestProvider_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "valid config",
			setup: func() {
				viper.Set("headers_username", "Remote-User")
				viper.Set("headers_groups", "Remote-Groups")
				viper.Set("headers_email", "Remote-Email")
				viper.Set("headers_name", "Remote-Name")
			},
			wantErr: false,
		},
		{
			name: "missing username header",
			setup: func() {
				viper.Set("headers_username", "")
				viper.Set("headers_groups", "Remote-Groups")
				viper.Set("headers_email", "Remote-Email")
				viper.Set("headers_name", "Remote-Name")
			},
			wantErr: true,
		},
		{
			name: "missing groups header",
			setup: func() {
				viper.Set("headers_username", "Remote-User")
				viper.Set("headers_groups", "")
				viper.Set("headers_email", "Remote-Email")
				viper.Set("headers_name", "Remote-Name")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := NewProvider(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseGroups(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "empty string",
			input:  "",
			expect: []string{},
		},
		{
			name:   "single group",
			input:  "group1",
			expect: []string{"group1"},
		},
		{
			name:   "multiple groups",
			input:  "group1,group2,group3",
			expect: []string{"group1", "group2", "group3"},
		},
		{
			name:   "groups with spaces",
			input:  "group1, group2 , group3",
			expect: []string{"group1", "group2", "group3"},
		},
		{
			name:   "groups with empty entries",
			input:  "group1,,group2,",
			expect: []string{"group1", "group2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGroups(tt.input)
			if len(result) != len(tt.expect) {
				t.Errorf("Expected %d groups, got %d", len(tt.expect), len(result))
				return
			}
			for i, expected := range tt.expect {
				if result[i] != expected {
					t.Errorf("Expected group[%d] = '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}