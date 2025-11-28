package libs

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"

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

// MainRoute is the main authentication handler that processes user authentication requests.
// It extracts user information from request headers, validates input, generates temporary passwords,
// optionally upserts the user to Elasticsearch, caches encrypted passwords, and returns basic auth credentials.
// The route supports caching to improve performance on repeated requests for the same user.
func MainRoute(c echo.Context) error {
	tracer := otel.Tracer("MainRoute")

	ctx, spanHeader := tracer.Start(c.Request().Context(), "Getting user information from request")

	spanHeader.AddEvent("Getting username from header")
	headerName := viper.GetString("headers_username")
	user := c.Request().Header.Get(headerName)

	if len(user) == 0 {
		err := errors.New("Header not provided: " + headerName)
		slog.ErrorContext(ctx, err.Error())
		spanHeader.RecordError(err)
		spanHeader.SetStatus(codes.Error, err.Error())
		response := ErrorResponse{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	if err := ValidateUsername(user); err != nil {
		slog.ErrorContext(ctx, "Invalid username format", slog.String("error", err.Error()))
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
	groupsHeader := c.Request().Header.Get(headerName)
	enableGroupWhitelist := viper.GetBool("enable_group_whitelist")
	var groupWhitelist []string
	if enableGroupWhitelist {
		groupWhitelist = viper.GetStringSlice("group_whitelist")
	}

	userGroups, err := ParseAndValidateGroups(groupsHeader, enableGroupWhitelist, groupWhitelist)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid group format", slog.String("error", err.Error()))
		response := ErrorResponse{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	if len(userGroups) == 0 && len(groupsHeader) == 0 {
		slog.DebugContext(ctx, "No groups provided in header")
	}

	ctx, spanCacheGet := tracer.Start(ctx, "cache get")
	cacheKey := "elastauth-" + EncodeForCacheKey(user)
	spanCacheGet.SetAttributes(attribute.String("user", user))
	spanCacheGet.SetAttributes(attribute.String("cacheKey", cacheKey))

	key := viper.GetString("secret_key")

	spanCacheGet.AddEvent("Getting password from cache")
	encryptedPasswordBase64, exists := cache.CacheInstance.Get(ctx, cacheKey)

	if exists {
		slog.DebugContext(ctx, "Cache hit", slog.String("cacheKey", cacheKey))
	} else {
		slog.DebugContext(ctx, "Cache miss", slog.String("cacheKey", cacheKey))
	}
	spanCacheGet.End()

	if !exists {
		ctx, spanCacheMiss := tracer.Start(ctx, "user access regeneration")

		roles := GetUserRoles(ctx, userGroups)

		userEmail := c.Request().Header.Get(viper.GetString("headers_email"))
		if len(userEmail) > 0 {
			if err := ValidateEmail(userEmail); err != nil {
				slog.ErrorContext(ctx, "Invalid email format", slog.String("error", err.Error()))
				response := ErrorResponse{
					Message: err.Error(),
					Code:    http.StatusBadRequest,
				}
				return c.JSON(http.StatusBadRequest, response)
			}
		}

		userName := c.Request().Header.Get(viper.GetString("headers_name"))
		if len(userName) > 0 {
			if err := ValidateName(userName); err != nil {
				slog.ErrorContext(ctx, "Invalid name format", slog.String("error", err.Error()))
				response := ErrorResponse{
					Message: err.Error(),
					Code:    http.StatusBadRequest,
				}
				return c.JSON(http.StatusBadRequest, response)
			}
		}

		spanCacheMiss.SetAttributes(attribute.String("userEmail", userEmail))
		spanCacheMiss.SetAttributes(attribute.String("userName", userName))

		spanCacheMiss.AddEvent("Generating temporary user password")
		password, err := GenerateTemporaryUserPassword(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to generate temporary user password", slog.Any("error", SanitizeForLogging(err)))
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Internal server error",
				Code:    http.StatusInternalServerError,
			})
		}

		spanCacheMiss.AddEvent("Encrypting temporary user password")
		encryptedPassword, err := Encrypt(ctx, password, key)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to encrypt password", slog.Any("error", SanitizeForLogging(err)))
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Internal server error",
				Code:    http.StatusInternalServerError,
			})
		}
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
				slog.ErrorContext(ctx, "Failed to initialize Elasticsearch client", slog.Any("error", SanitizeForLogging(err)))
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Message: "Internal server error",
					Code:    http.StatusInternalServerError,
				})
			}

			spanCacheMiss.AddEvent("Upserting user in Elasticsearch")
			err = UpsertUser(ctx, user, elasticsearchUser)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to upsert user in Elasticsearch", slog.Any("error", SanitizeForLogging(err)))
				return c.JSON(http.StatusInternalServerError, ErrorResponse{
					Message: "Internal server error",
					Code:    http.StatusInternalServerError,
				})
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
		slog.DebugContext(ctx, "Extending cache TTL", slog.String("user", user), slog.Duration("currentTTL", itemCacheDuration), slog.String("configuredTTL", viper.GetString("cache_expire")))
		cache.CacheInstance.ExtendTTL(ctx, cacheKey, encryptedPasswordBase64)
	}
	spanItemCache.End()

	ctx, spanDecrypt := tracer.Start(ctx, "password decryption")
	spanDecrypt.AddEvent("Decrypting password")

	decryptedPasswordBase64, err := base64.URLEncoding.DecodeString(encryptedPasswordBase64.(string))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decode password from cache", slog.Any("error", SanitizeForLogging(err)))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}

	decryptedPassword, err := Decrypt(ctx, string(decryptedPasswordBase64), key)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decrypt password", slog.Any("error", SanitizeForLogging(err)))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Internal server error",
			Code:    http.StatusInternalServerError,
		})
	}
	spanDecrypt.End()

	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))

	return c.NoContent(http.StatusOK)
}

// HealthRoute responds to health checks with a JSON response containing the application status.
// This endpoint is typically used by load balancers or monitoring systems to verify the application is running.
func HealthRoute(c echo.Context) error {
	tracer := otel.Tracer("HealthRoute")
	_, span := tracer.Start(c.Request().Context(), "response")
	defer span.End()

	response := HealthResponse{
		Status: "OK",
	}

	return c.JSON(http.StatusOK, response)
}

// ConfigRoute returns the application's configuration for default roles and group-to-role mappings.
// This endpoint allows clients to discover which roles users will receive based on their groups.
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
