package libs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	tracer := otel.Tracer("MainRoute")

	ctx, spanHeader := tracer.Start(c.Request().Context(), "Getting user information from request")

	spanHeader.AddEvent("Getting username from header")
	headerName := viper.GetString("headers_username")
	user := c.Request().Header.Get(headerName)

	if len(user) == 0 {
		err := errors.New("Header not provided: " + headerName)
		slog.ErrorCtx(ctx, err.Error())
		spanHeader.RecordError(err)
		spanHeader.SetStatus(codes.Error, err.Error())
		response := ErrorResponse{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, response)
	}
	spanHeader.End()

	headerName = viper.GetString("headers_groups")
	userGroups := strings.Split(c.Request().Header.Get(headerName), ",")

	if len(userGroups) == 0 {
		errorMessage := "Header not provided: " + headerName
		slog.ErrorCtx(ctx, errorMessage)
	}

	ctx, spanCacheGet := tracer.Start(ctx, "cache get")
	cacheKey := "elastauth-" + user
	spanCacheGet.SetAttributes(attribute.String("user", user))
	spanCacheGet.SetAttributes(attribute.String("cacheKey", cacheKey))

	key := viper.GetString("secret_key")

	spanCacheGet.AddEvent("Getting password from cache")
	encryptedPasswordBase64, exists := cache.CacheInstance.Get(ctx, cacheKey)

	if exists {
		slog.DebugCtx(ctx, "Cache hit", slog.String("cacheKey", cacheKey))
	} else {
		slog.DebugCtx(ctx, "Cache miss", slog.String("cacheKey", cacheKey))
	}
	spanCacheGet.End()

	if !exists {
		ctx, spanCacheMiss := tracer.Start(ctx, "user access regeneration")

		roles := GetUserRoles(ctx, userGroups)

		userEmail := c.Request().Header.Get(viper.GetString("headers_email"))
		userName := c.Request().Header.Get(viper.GetString("headers_name"))
		spanCacheMiss.SetAttributes(attribute.String("userEmail", userEmail))
		spanCacheMiss.SetAttributes(attribute.String("userName", userName))

		spanCacheMiss.AddEvent("Generating temporary user password")
		password, err := GenerateTemporaryUserPassword(ctx)
		if err != nil {
			slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
			return c.JSON(http.StatusInternalServerError, err)
		}

		spanCacheMiss.AddEvent("Encrypting temporary user password")
		encryptedPassword := Encrypt(ctx, password, key)
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
				ctx,
				viper.GetString("elasticsearch_host"),
				viper.GetString("elasticsearch_username"),
				viper.GetString("elasticsearch_password"),
			)
			if err != nil {
				slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
				return c.JSON(http.StatusInternalServerError, err)
			}

			spanCacheMiss.AddEvent("Upserting user in Elasticsearch")
			err = UpsertUser(ctx, user, elasticsearchUser)
			if err != nil {
				slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
				return c.JSON(http.StatusInternalServerError, err)
			}
		}

		spanCacheMiss.AddEvent("Setting cache item")
		cache.CacheInstance.Set(ctx, cacheKey, encryptedPasswordBase64)
		spanCacheMiss.End()
	}

	ctx, spanItemCache := tracer.Start(ctx, "handling item cache")
	spanItemCache.SetAttributes(attribute.String("cacheKey", cacheKey))

	itemCacheDuration, _ := cache.CacheInstance.GetItemTTL(ctx, cacheKey)

	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cache.CacheInstance.GetTTL(ctx) {
		slog.DebugCtx(ctx, fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("cache_expire")))
		cache.CacheInstance.ExtendTTL(ctx, cacheKey, encryptedPasswordBase64)
	}
	spanItemCache.End()

	ctx, spanDecrypt := tracer.Start(ctx, "password decryption")
	spanDecrypt.AddEvent("Decrypting password")

	decryptedPasswordBase64, _ := base64.URLEncoding.DecodeString(encryptedPasswordBase64.(string))

	decryptedPassword := Decrypt(ctx, string(decryptedPasswordBase64), key)
	spanDecrypt.End()

	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))

	return c.NoContent(http.StatusOK)
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
