package libs

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

// The HealthResponse type is a struct in Go that contains a single field called Status, which is a
// string that will be represented as "status" in JSON.
// @property {string} Status - The `Status` property is a string field that represents the status of a
// health response. It is tagged with `json:"status"` which indicates that when this struct is
// serialized to JSON, the field name will be "status".
type HealthResponse struct {
	Status string `json:"status"`
}

// The type `ErrorResponse` is a struct that contains a message and code for error responses in Go.
// @property {string} Message - Message is a string property that represents the error message that
// will be returned in the response when an error occurs.
// @property {int} Code - The `Code` property is an integer that represents an error code. It is used
// to identify the type of error that occurred. For example, a code of 404 might indicate that a
// requested resource was not found, while a code of 500 might indicate a server error.
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// The configResponse type contains default roles and group mappings in a map format.
// @property {[]string} DefaultRoles - DefaultRoles is a property of the configResponse struct that is
// a slice of strings representing the default roles assigned to users who do not have any specific
// roles assigned to them. These roles can be used to grant basic permissions to all users in the
// system.
// @property GroupMappings - `GroupMappings` is a property of the `configResponse` struct that is a map
// of strings to slices of strings. It is used to map groups to roles in the application. Each key in
// the map represents a group, and the corresponding value is a slice of roles that are associated with
// that
type configResponse struct {
	DefaultRoles  []string            `json:"default_roles"`
	GroupMappings map[string][]string `json:"group_mappings"`
}

// This function handles the main route of a web application, authenticating users and caching their
// encrypted passwords.
func MainRoute(c echo.Context) error {
	return handleMainRequest(c)
}

// The function returns a JSON response with a "OK" status for a health route in a Go application.
func HealthRoute(c echo.Context) error {
	tracer := otel.Tracer("HealthRoute")
	_, span := tracer.Start(c.Request().Context(), "response")
	defer span.End()

	response := HealthResponse{
		Status: "OK",
	}

	return c.JSON(http.StatusOK, response)
}

// The function returns a JSON response containing default roles and group mappings from a
// configuration file.
func ConfigRoute(c echo.Context) error {
	tracer := otel.Tracer("ConfigRoute")
	_, span := tracer.Start(c.Request().Context(), "response")
	defer span.End()

	response := configResponse{
		DefaultRoles:  viper.GetStringSlice("default_roles"),
		GroupMappings: viper.GetStringMapStringSlice("group_mappings"),
	}
	return c.JSON(http.StatusOK, response)
}
