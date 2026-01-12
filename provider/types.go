package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// AuthProvider defines the interface that all authentication providers must implement
type AuthProvider interface {
	// GetUser extracts user information from the authentication request
	GetUser(ctx context.Context, req *AuthRequest) (*UserInfo, error)
	
	// Type returns the provider type identifier
	Type() string
	
	// Validate checks if the provider configuration is valid
	Validate() error
}

// AuthRequest wraps the HTTP request with helper methods for different auth mechanisms
type AuthRequest struct {
	*http.Request
}

// GetHeader returns the value of the specified header
func (r *AuthRequest) GetHeader(key string) string {
	return r.Header.Get(key)
}

// GetCookie returns the named cookie provided in the request
func (r *AuthRequest) GetCookie(key string) (*http.Cookie, error) {
	return r.Request.Cookie(key)
}

// GetQueryParam returns the value of the specified query parameter
func (r *AuthRequest) GetQueryParam(key string) string {
	return r.URL.Query().Get(key)
}

// GetBearerToken extracts the bearer token from the Authorization header
func (r *AuthRequest) GetBearerToken() (string, error) {
	auth := r.GetHeader("Authorization")
	if auth == "" {
		return "", fmt.Errorf("authorization header not found")
	}
	
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(auth, bearerPrefix) {
		return "", fmt.Errorf("authorization header does not contain bearer token")
	}
	
	return auth[len(bearerPrefix):], nil
}

// UserInfo represents standardized user information from any provider
type UserInfo struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Groups   []string `json:"groups"`
	FullName string   `json:"full_name"`
}