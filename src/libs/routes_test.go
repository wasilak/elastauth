package libs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	response := HealthResponse{
		Status: "OK",
	}

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, response, rec.Body.String())
}
