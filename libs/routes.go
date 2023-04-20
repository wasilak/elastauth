package libs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type configResponse struct {
	DefaultRoles  []string            `json:"default_roles"`
	GroupMappings map[string][]string `json:"group_mappings"`
}

func MainRoute(c echo.Context) error {
	headerName := viper.GetString("headers_username")
	user := c.Request().Header.Get(headerName)

	if len(user) == 0 {
		errorMessage := "Header not provided: " + headerName
		log.Error(errorMessage)
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
		log.Error(errorMessage)
	}

	cacheKey := "elastauth-" + user

	key := viper.GetString("secret_key")

	encryptedPasswordBase64, exists := cache.CacheInstance.Get(cacheKey)

	if exists {
		log.Debug(fmt.Sprintf("Cache hit: %s", cacheKey))
	} else {
		log.Debug(fmt.Sprintf("Cache miss: %s", cacheKey))
	}

	if !exists {
		roles := GetUserRoles(userGroups)

		userEmail := c.Request().Header.Get(viper.GetString("headers_email"))
		userName := c.Request().Header.Get(viper.GetString("headers_name"))

		password, err := GenerateTemporaryUserPassword()
		if err != nil {
			log.Error(err)
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
				log.Error(err)
				return c.JSON(http.StatusInternalServerError, err)
			}

			err = UpsertUser(user, elasticsearchUser)
			if err != nil {
				log.Error(err)
				return c.JSON(http.StatusInternalServerError, err)
			}
		}

		cache.CacheInstance.Set(cacheKey, encryptedPasswordBase64)
	}

	itemCacheDuration, _ := cache.CacheInstance.GetItemTTL(cacheKey)

	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cache.CacheInstance.GetTTL() {
		log.Debug(fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("cache_expire")))
		cache.CacheInstance.ExtendTTL(cacheKey, encryptedPasswordBase64)
	}

	decryptedPasswordBase64, _ := base64.URLEncoding.DecodeString(encryptedPasswordBase64.(string))

	decryptedPassword := Decrypt(string(decryptedPasswordBase64), key)

	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))

	return c.NoContent(http.StatusOK)
}

func HealthRoute(c echo.Context) error {
	response := HealthResponse{
		Status: "OK",
	}

	return c.JSON(http.StatusOK, response)
}

func ConfigRoute(c echo.Context) error {
	response := configResponse{
		DefaultRoles:  viper.GetStringSlice("default_roles"),
		GroupMappings: viper.GetStringMapStringSlice("group_mappings"),
	}
	return c.JSON(http.StatusOK, response)
}
