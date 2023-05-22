package libs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"golang.org/x/exp/slog"
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
	headerName := viper.GetString("headers_username")
	user := c.Request().Header.Get(headerName)

	if len(user) == 0 {
		errorMessage := "Header not provided: " + headerName
		slog.Error(errorMessage)
		response := ErrorResponse{
			Message: errorMessage,
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	headerName = viper.GetString("headers_groups")
	userGroups := strings.Split(c.Request().Header.Get(headerName), ",")

	if len(userGroups) == 0 {
		errorMessage := "Header not provided: " + headerName
		slog.Error(errorMessage)
	}

	cacheKey := "elastauth-" + user

	key := viper.GetString("secret_key")

	encryptedPasswordBase64, exists := cache.CacheInstance.Get(cacheKey)

	if exists {
		slog.Debug("Cache hit", slog.String("cacheKey", cacheKey))
	} else {
		slog.Debug("Cache miss", slog.String("cacheKey", cacheKey))
	}

	if !exists {
		roles := GetUserRoles(userGroups)

		userEmail := c.Request().Header.Get(viper.GetString("headers_email"))
		userName := c.Request().Header.Get(viper.GetString("headers_name"))

		password, err := GenerateTemporaryUserPassword()
		if err != nil {
			slog.Error("Error", slog.Any("message", err))
			return c.JSON(http.StatusInternalServerError, err)
		}
		encryptedPassword := Encrypt(password, key)
		encryptedPasswordBase64 = string(base64.URLEncoding.EncodeToString([]byte(encryptedPassword)))

		elasticsearchUserMetadata := ElasticsearchUserMetadata{
			Groups: userGroups,
		}

		elasticsearchUser := ElasticsearchUser{
			Password: password,
			Enabled:  true,
			Email:    userEmail,
			FullName: userName,
			Roles:    roles,
			Metadata: elasticsearchUserMetadata,
		}

		if !viper.GetBool("elasticsearch_dry_run") {
			err := initElasticClient(
				viper.GetString("elasticsearch_host"),
				viper.GetString("elasticsearch_username"),
				viper.GetString("elasticsearch_password"),
			)
			if err != nil {
				slog.Error("Error", slog.Any("message", err))
				return c.JSON(http.StatusInternalServerError, err)
			}

			err = UpsertUser(user, elasticsearchUser)
			if err != nil {
				slog.Error("Error", slog.Any("message", err))
				return c.JSON(http.StatusInternalServerError, err)
			}
		}

		cache.CacheInstance.Set(cacheKey, encryptedPasswordBase64)
	}

	itemCacheDuration, _ := cache.CacheInstance.GetItemTTL(cacheKey)

	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cache.CacheInstance.GetTTL() {
		slog.Debug(fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("cache_expire")))
		cache.CacheInstance.ExtendTTL(cacheKey, encryptedPasswordBase64)
	}

	decryptedPasswordBase64, _ := base64.URLEncoding.DecodeString(encryptedPasswordBase64.(string))

	decryptedPassword := Decrypt(string(decryptedPasswordBase64), key)

	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))

	return c.NoContent(http.StatusOK)
}

// The function returns a JSON response with a "OK" status for a health route in a Go application.
func HealthRoute(c echo.Context) error {
	response := HealthResponse{
		Status: "OK",
	}

	return c.JSON(http.StatusOK, response)
}

// The function returns a JSON response containing default roles and group mappings from a
// configuration file.
func ConfigRoute(c echo.Context) error {
	response := configResponse{
		DefaultRoles:  viper.GetStringSlice("default_roles"),
		GroupMappings: viper.GetStringMapStringSlice("group_mappings"),
	}
	return c.JSON(http.StatusOK, response)
}
