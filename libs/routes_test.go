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

func TestConfigRoute(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/config")

	viperDefaultRolesMock := []string{"your_default_kibana_role"}
	viper.Set("default_roles", viperDefaultRolesMock)

	viperMappingsMock := map[string][]string{
		"your_ad_group": {"your_kibana_role"},
	}
	viper.Set("group_mappings", viperMappingsMock)

	response := "{\"default_roles\":[\"your_default_kibana_role\"],\"group_mappings\":{\"your_ad_group\":[\"your_kibana_role\"]}}\n"

	// Assertions
	if assert.NoError(t, ConfigRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, response, rec.Body.String())
	}
}
