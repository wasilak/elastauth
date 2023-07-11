package libs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/exp/slog"
)

// The line `var tracerMainRoute = otel.Tracer("MainRoute")` is creating a new instance of an
// OpenTelemetry tracer named "MainRoute". This tracer can be used to instrument and trace specific
// sections of code for monitoring and observability purposes.
var tracerMainRoute = otel.Tracer("MainRoute")

// The function handles a main request by retrieving user information from headers, caching and
// encrypting the data, and setting the Authorization header in the response.
// The function handleMainRequest takes in an echo.Context and returns an error.
// It handles the main request by performing the following steps:
func handleMainRequest(c echo.Context) error {

	// Get the context from the echo.Context object
	ctx := c.Request().Context()

	// Call getInfoFromHeaders to extract user information from headers
	// This function returns a new context, user, userGroups, and an error
	ctx, user, userGroups, err := getInfoFromHeaders(ctx, c)
	if err != nil {
		// If there is an error, create an ErrorResponse with the message and code
		response := ErrorResponse{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		// Return the ErrorResponse as JSON with a status code of 400 (Bad Request)
		return c.JSON(http.StatusBadRequest, response)
	}

	// Get the secret key from the configuration
	key := viper.GetString("secret_key")

	// Create a cacheKey using the user
	cacheKey := "elastauth-" + user

	// Check if the user's data exists in the cache
	exists, encryptedPasswordBase64 := handleCachedData(ctx, cacheKey, user)

	// If the user's data does not exist, regenerate it and store it in the cache
	if !exists {
		err = regenerateCache(ctx, c, cacheKey, key, user, userGroups)
		if err != nil {
			// If there is an error regenerating the cache, return it as JSON with a status code of 500 (Internal Server Error)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	// Extend the expiration time of the user's data in the cache
	extendingDataCache(ctx, cacheKey, user, encryptedPasswordBase64)

	// Decrypt the user's password using the secret key
	decryptedPassword, err := decryptingPassword(ctx, key, encryptedPasswordBase64)
	if err != nil {
		// If there is an error decrypting the password, return it as JSON with a status code of 500 (Internal Server Error)
		c.JSON(http.StatusInternalServerError, err)
	}

	// Set the Authorization header in the response with basic authentication using the user and decrypted password
	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))

	// Return a response with a status code of 200 (OK) and no content
	return c.NoContent(http.StatusOK)
}

// The function `getInfoFromHeaders` retrieves user information from the request headers in an Echo
// context.
// The function getInfoFromHeaders takes in a context.Context and an echo.Context object and returns a new context, user, userGroups, and an error.
// It retrieves user information from the request headers by performing the following steps:
func getInfoFromHeaders(ctx context.Context, c echo.Context) (context.Context, string, []string, error) {

	// Start a tracing span named "Getting user information from request"
	ctx, span := tracerMainRoute.Start(ctx, "Getting user information from request")

	// Add an event to the span indicating that the username is being retrieved from the header
	span.AddEvent("Getting username from header")

	// Get the header name for the username from the configuration
	headerName := viper.GetString("headers_username")

	// Get the value of the username header from the request
	user := c.Request().Header.Get(headerName)

	var userGroups []string

	// If the length of the username is 0, it means the header was not provided
	if len(user) == 0 {
		// Create an error with a descriptive message
		err := errors.New("Header not provided: " + headerName)
		// Log the error using slog.ErrorCtx
		slog.ErrorCtx(ctx, err.Error())
		// Record the error in the tracing span
		span.RecordError(err)
		// Set the status of the span to indicate an error occurred
		span.SetStatus(codes.Error, err.Error())
		// Return the current context, empty user, empty userGroups, and the error
		return ctx, "", userGroups, err
	}

	// Get the header name for the user groups from the configuration
	headerName = viper.GetString("headers_groups")

	// Get the value of the user groups header from the request and split it into individual groups
	userGroups = strings.Split(c.Request().Header.Get(headerName), ",")

	// If no user groups are provided, log an error and return the current context, empty user, userGroups, and the error
	if len(userGroups) == 0 {
		err := errors.New("Header not provided: " + headerName)
		slog.ErrorCtx(ctx, err.Error())
		return ctx, "", userGroups, err
	}

	// End the tracing span
	span.End()

	// Return the updated context, user, userGroups, and nil error
	return ctx, user, userGroups, nil
}

// The function `handleCachedData` retrieves data from a cache using a cache key and user identifier,
// and returns a boolean indicating if the data exists in the cache and the encrypted password in
// base64 format.
// The function handleCachedData takes in a context.Context, a cache key, and a user string, and returns a boolean indicating whether the data is found in the cache and the encrypted password.
func handleCachedData(ctx context.Context, cacheKey, user string) (bool, interface{}) {

	// Start a tracing span named "cache get"
	ctx, span := tracerMainRoute.Start(ctx, "cache get")

	// Set attributes on the span to include the user and cacheKey
	span.SetAttributes(attribute.String("user", user))
	span.SetAttributes(attribute.String("cacheKey", cacheKey))

	// Add an event to the span indicating that the password is being retrieved from the cache
	span.AddEvent("Getting password from cache")

	// Retrieve the encrypted password from the cache using the cacheKey
	encryptedPasswordBase64, exists := cache.CacheInstance.Get(ctx, cacheKey)

	// If the data exists in the cache, log a debug message indicating a cache hit
	if exists {
		slog.DebugCtx(ctx, "Cache hit", slog.String("cacheKey", cacheKey))
	} else {
		// If the data does not exist in the cache, log a debug message indicating a cache miss
		slog.DebugCtx(ctx, "Cache miss", slog.String("cacheKey", cacheKey))
	}

	// End the tracing span
	span.End()

	// Return the boolean indicating whether the data exists in the cache and the encrypted password
	return exists, encryptedPasswordBase64
}

// The `regenerateCache` function generates a temporary user password, encrypts it, and stores it in a
// cache.
func regenerateCache(ctx context.Context, c echo.Context, cacheKey string, key string, user string, userGroups []string) error {
	// Start a tracing span named "user access regeneration" and store the context and span in variables ctx and spanCacheMiss.
	ctx, spanCacheMiss := tracerMainRoute.Start(ctx, "user access regeneration")

	// Get user roles by calling the GetUserRoles function passing in the context and userGroups.
	roles := GetUserRoles(ctx, userGroups)

	// Get the user's email and name from the request headers.
	userEmail := c.Request().Header.Get(viper.GetString("headers_email"))
	userName := c.Request().Header.Get(viper.GetString("headers_name"))

	// Set attributes on the span to include the userEmail and userName.
	spanCacheMiss.SetAttributes(attribute.String("userEmail", userEmail))
	spanCacheMiss.SetAttributes(attribute.String("userName", userName))

	// Add an event to the span indicating that a temporary user password is being generated.
	spanCacheMiss.AddEvent("Generating temporary user password")

	// Generate a temporary user password by calling the GenerateTemporaryUserPassword function passing in the context.
	password, err := GenerateTemporaryUserPassword(ctx)
	if err != nil {
		// If there is an error generating the temporary password, log the error and return it.
		slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
		return err
	}

	// Add an event to the span indicating that the temporary user password is being encrypted.
	spanCacheMiss.AddEvent("Encrypting temporary user password")

	// Encrypt the temporary user password by calling the Encrypt function passing in the context, password, and key.
	encryptedPassword, err := Encrypt(ctx, password, key)
	if err != nil {
		// If there is an error encrypting the password, log the error and return it.
		slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
		return err
	}

	// Log the encrypted password.
	slog.Info(encryptedPassword)

	// Encode the encrypted password to base64.
	encryptedPasswordBase64 := string(base64.URLEncoding.EncodeToString([]byte(encryptedPassword)))

	// Create an ElasticsearchUserMetadata struct with the userGroups.
	elasticsearchUserMetadata := ElasticsearchUserMetadata{
		Groups: userGroups,
	}

	// Create an ElasticsearchUser struct with the password, email, name, roles, and metadata.
	elasticsearchUser := ElasticsearchUser{
		Password: password,
		Enabled:  true,
		Email:    userEmail,
		FullName: userName,
		Roles:    roles,
		Metadata: elasticsearchUserMetadata,
	}

	// If the elasticsearch_dry_run configuration is false, initialize the Elastic client and upsert the user in Elasticsearch.
	if !viper.GetBool("elasticsearch_dry_run") {
		err := initElasticClient(
			ctx,
			viper.GetString("elasticsearch_host"),
			viper.GetString("elasticsearch_username"),
			viper.GetString("elasticsearch_password"),
		)
		if err != nil {
			// If there is an error initializing the Elastic client, log the error and return it.
			slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Add an event to the span indicating that the user is being upserted in Elasticsearch.
		spanCacheMiss.AddEvent("Upserting user in Elasticsearch")

		// Upsert the user in Elasticsearch by calling the UpsertUser function passing in the context, user, and elasticsearchUser.
		err = UpsertUser(ctx, user, elasticsearchUser)
		if err != nil {
			// If there is an error upserting the user in Elasticsearch, log the error and return it.
			slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	// Add an event to the span indicating that the cache item is being set.
	spanCacheMiss.AddEvent("Setting cache item")

	// Set the cache item by calling the Set method on CacheInstance passing in the context, cacheKey, and encryptedPasswordBase64.
	cache.CacheInstance.Set(ctx, cacheKey, encryptedPasswordBase64)

	// End the tracing span.
	spanCacheMiss.End()

	// Return nil indicating success.
	return nil
}

// The function extends the time-to-live (TTL) of an item in the cache if certain conditions are met.
func extendingDataCache(ctx context.Context, cacheKey string, user string, encryptedPasswordBase64 interface{}) {
	// Start a tracing span named "handling item cache" and store the context and span in variables ctx and spanItemCache.
	ctx, spanItemCache := tracerMainRoute.Start(ctx, "handling item cache")

	// Set an attribute on the span to include the cacheKey.
	spanItemCache.SetAttributes(attribute.String("cacheKey", cacheKey))

	// Get the time to live (TTL) of the item in the cache by calling the GetItemTTL function on CacheInstance.
	itemCacheDuration, _ := cache.CacheInstance.GetItemTTL(ctx, cacheKey)

	// Check if the extend_cache configuration is true, the itemCacheDuration is greater than 0, and less than the overall TTL of the cache.
	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cache.CacheInstance.GetTTL(ctx) {
		// If the conditions are met, log a debug message indicating that the cache TTL is being extended.
		slog.DebugCtx(ctx, fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("cache_expire")))

		// Extend the TTL of the cache item by calling the ExtendTTL method on CacheInstance passing in the context, cacheKey, and encryptedPasswordBase64.
		cache.CacheInstance.ExtendTTL(ctx, cacheKey, encryptedPasswordBase64)
	}

	// End the tracing span.
	spanItemCache.End()
}

// The function decrypts an encrypted password using a provided key.
func decryptingPassword(ctx context.Context, key string, encryptedPasswordBase64 interface{}) (string, error) {
	// Start a span for password decryption
	ctx, spanDecrypt := tracerMainRoute.Start(ctx, "password decryption")
	spanDecrypt.AddEvent("Decrypting password")

	encryptedPassword, ok := encryptedPasswordBase64.(string)
	if !ok {
		return "", fmt.Errorf("encrypted password is not a string")
	}

	// Decode the encrypted password from base64
	decryptedPasswordBase64, err := base64.URLEncoding.DecodeString(encryptedPassword)
	if err != nil {
		// Log any errors that occur during base63 encryption/decryption
		slog.ErrorCtx(ctx, "Error", slog.Any("message", err))
		return "", err
	}

	// Decrypt the password using the provided key
	decryptedPassword, err := Decrypt(ctx, string(decryptedPasswordBase64), key)
	if err != nil {
		// Log any errors that occur during decryption
		slog.ErrorCtx(ctx, "error decrypting password", slog.Any("message", err))
		return "", err
	}

	// End the span for password decryption
	spanDecrypt.End()

	return decryptedPassword, nil
}
