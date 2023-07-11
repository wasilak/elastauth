package libs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHealthRoute(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	response := "{\"status\":\"OK\"}\n"

	// Assertions
	if assert.NoError(t, HealthRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, response, rec.Body.String())
	}
}

// TestConfigRoute is a test function for the ConfigRoute handler.
func TestConfigRoute(t *testing.T) {
	// Setup
	// Create a new instance of Echo framework.
	e := echo.New()

	// Create a new HTTP request with method GET and path "/".
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create a new HTTP response recorder.
	rec := httptest.NewRecorder()

	// Create a new context with the request and response recorder.
	c := e.NewContext(req, rec)

	// Set the path to "/config" in the context.
	c.SetPath("/config")

	// Mock the default roles using viper package.
	viperDefaultRolesMock := []string{"your_default_kibana_role"}
	viper.Set("default_roles", viperDefaultRolesMock)

	// Mock the group mappings using viper package.
	viperMappingsMock := map[string][]string{
		"your_ad_group": {"your_kibana_role"},
	}
	viper.Set("group_mappings", viperMappingsMock)

	// Define the expected response as a JSON string.
	response := "{\"default_roles\":[\"your_default_kibana_role\"],\"group_mappings\":{\"your_ad_group\":[\"your_kibana_role\"]}}\n"

	// Assertions
	// Call the ConfigRoute handler function and check if there are no errors.
	if assert.NoError(t, ConfigRoute(c)) {
		// Assert that the HTTP status code is 200 OK.
		assert.Equal(t, http.StatusOK, rec.Code)

		// Assert that the response body matches the expected response.
		assert.Equal(t, response, rec.Body.String())
	}
}
