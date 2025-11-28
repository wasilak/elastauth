package libs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitElasticClient_Success(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		auth := r.Header.Get("Authorization")
		assert.NotEmpty(t, auth)

		response := map[string]interface{}{
			"name":         "elasticsearch",
			"cluster_name": "elasticsearch",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	err := initElasticClient(ctx, server.URL, "user", "password")

	assert.NoError(t, err)
	assert.NotNil(t, elasticsearchConnectionDetails)
	assert.Equal(t, server.URL, elasticsearchConnectionDetails.URL)
	assert.Equal(t, "user", elasticsearchConnectionDetails.Username)
	assert.Equal(t, "password", elasticsearchConnectionDetails.Password)
}

func TestInitElasticClient_ConnectionRefused(t *testing.T) {
	ctx := context.Background()

	err := initElasticClient(ctx, "http://localhost:99999", "user", "password")

	assert.Error(t, err)
}

func TestInitElasticClient_InvalidResponse(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json {"))
	}))
	defer server.Close()

	err := initElasticClient(ctx, server.URL, "user", "password")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode Elasticsearch response")
}

func TestInitElasticClient_InvalidURL(t *testing.T) {
	ctx := context.Background()

	err := initElasticClient(ctx, "ht!tp://invalid", "user", "password")

	assert.Error(t, err)
}

func TestUpsertUser_Success(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/_security/user/testuser")

		auth := r.Header.Get("Authorization")
		assert.NotEmpty(t, auth)

		contentType := r.Header.Get("Content-Type")
		assert.Equal(t, "application/json", contentType)

		body := ElasticsearchUser{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "testuser@example.com", body.Email)
		assert.Equal(t, "Test User", body.FullName)
		assert.True(t, body.Enabled)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"created": true})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "testuser@example.com",
		Password: "securePassword123",
		Metadata: ElasticsearchUserMetadata{
			Groups: []string{"admin", "users"},
		},
		FullName: "Test User",
		Roles:    []string{"admin", "superuser"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.NoError(t, err)
}

func TestUpsertUser_InvalidJSON(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json {"))
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "testuser@example.com",
		Password: "securePassword123",
		FullName: "Test User",
		Roles:    []string{"admin"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode Elasticsearch response")
}

func TestUpsertUser_HTTPError(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Internal server error"})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "testuser@example.com",
		Password: "securePassword123",
		FullName: "Test User",
		Roles:    []string{"admin"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed with status 500")
}

func TestUpsertUser_NetworkError(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://localhost:99999",
		Username: "user",
		Password: "password",
	}

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "testuser@example.com",
		Password: "securePassword123",
		FullName: "Test User",
		Roles:    []string{"admin"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.Error(t, err)
}

func TestUpsertUser_WithAllMetadata(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		body := ElasticsearchUser{}
		json.NewDecoder(r.Body).Decode(&body)

		assert.Equal(t, "admin@example.com", body.Email)
		assert.Equal(t, "Admin User", body.FullName)
		assert.True(t, body.Enabled)
		assert.Equal(t, []string{"admin", "users"}, body.Metadata.Groups)
		assert.Equal(t, []string{"admin", "superuser", "viewer"}, body.Roles)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"created": true})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "admin@example.com",
		Password: "securePassword123",
		Metadata: ElasticsearchUserMetadata{
			Groups: []string{"admin", "users"},
		},
		FullName: "Admin User",
		Roles:    []string{"admin", "superuser", "viewer"},
	}

	err := UpsertUser(ctx, "adminuser", user)

	assert.NoError(t, err)
}

func TestUpsertUser_EmptyEmail(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"created": true})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "",
		Password: "securePassword123",
		FullName: "Test User",
		Roles:    []string{"user"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.NoError(t, err)
}

func TestUpsertUser_ForbiddenStatus(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Permission denied"})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "testuser@example.com",
		Password: "securePassword123",
		FullName: "Test User",
		Roles:    []string{"admin"},
	}

	err := UpsertUser(ctx, "testuser", user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed with status 403")
}

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "standard credentials",
			username: "user",
			password: "pass",
			expected: "dXNlcjpwYXNz",
		},
		{
			name:     "credentials with special chars",
			username: "user@domain.com",
			password: "p@ss:word",
			expected: basicAuth("user@domain.com", "p@ss:word"),
		},
		{
			name:     "empty password",
			username: "user",
			password: "",
			expected: basicAuth("user", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basicAuth(tt.username, tt.password)
			assert.NotEmpty(t, result)

			if tt.name == "standard credentials" {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestElasticsearchUserMarshaling(t *testing.T) {
	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "test@example.com",
		Password: "password",
		Metadata: ElasticsearchUserMetadata{
			Groups: []string{"admin", "users"},
		},
		FullName: "Test User",
		Roles:    []string{"admin", "superuser"},
	}

	jsonBytes, err := json.Marshal(user)
	assert.NoError(t, err)

	unmarshaled := ElasticsearchUser{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, user.Email, unmarshaled.Email)
	assert.Equal(t, user.FullName, unmarshaled.FullName)
	assert.Equal(t, user.Enabled, unmarshaled.Enabled)
	assert.Equal(t, user.Roles, unmarshaled.Roles)
	assert.Equal(t, user.Metadata.Groups, unmarshaled.Metadata.Groups)
}

func TestUpsertUser_SpecialCharactersInUsername(t *testing.T) {
	ctx := context.Background()

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      "http://elasticsearch:9200",
		Username: "user",
		Password: "password",
	}

	capturedUsername := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUsername = r.URL.Path

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"created": true})
	}))
	defer server.Close()

	elasticsearchConnectionDetails.URL = server.URL

	user := ElasticsearchUser{
		Enabled:  true,
		Email:    "test@example.com",
		Password: "password",
		FullName: "Test User",
		Roles:    []string{"user"},
	}

	err := UpsertUser(ctx, "user.name-123_test", user)

	assert.NoError(t, err)
	assert.Contains(t, capturedUsername, "user.name-123_test")
}
