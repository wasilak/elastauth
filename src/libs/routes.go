package libs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	log.Debug(c.Request().Header)

	headerName := "Remote-User"
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

	headerName = "Remote-Groups"
	userGroups := strings.Split(c.Request().Header.Get(headerName), ",")

	if len(userGroups) == 0 {
		errorMessage := "Header not provided: " + headerName
		log.Error(errorMessage)
	}

	cacheDuration, err := time.ParseDuration(viper.GetString("redis_expire_seconds"))
	if err != nil {
		log.Fatal(err)
	}

	var cacheInstance cache.CacheInterface
	if viper.GetString("cache_type") == "memory" {
		cacheInstance = &cache.RedisCache{
			Address: fmt.Sprintf("%s:%s", viper.GetString("redis_host"), viper.GetString("redis_port")),
			DB:      viper.GetInt("redis_db"),
		}
	} else if viper.GetString("cache_type") == "memory" {
		cacheInstance = &cache.RistrettoCache{}
	} else {
		log.Fatal("No cache_type selected or cache type is invalid")
	}

	cacheInstance.Init(cacheDuration)

	cacheKey := "elastauth-" + user

	encryptedPasswordBase64, exists := cacheInstance.Get(cacheKey)

	key := viper.GetString("secret_key")

	if exists {
		roles := GetUserRoles(userGroups)

		userEmail := c.Request().Header.Get("Remote-Email")
		userName := c.Request().Header.Get("Remote-Name")

		initElasticClient(
			viper.GetString("elasticsearch_host"),
			viper.GetString("elasticsearch_username"),
			viper.GetString("elasticsearch_password"),
		)

		password := GenerateTemporaryUserPassword()

		encryptedPassword := Encrypt(password, key)

		encryptedPasswordBase64 = string(base64.URLEncoding.EncodeToString([]byte(encryptedPassword)))

		log.Debug(encryptedPasswordBase64)

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

		UpsertUser(user, elasticsearchUser)

		cacheInstance.Set(cacheKey, encryptedPasswordBase64)
	} else if err != nil {
		panic(err)
	}

	itemCacheDuration, _ := cacheInstance.GetTTL(cacheKey)

	log.Debug(cacheDuration, itemCacheDuration)

	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cacheDuration {
		log.Debug(fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("redis_expire_seconds")))
		cacheInstance.ExtendTTL(cacheKey, encryptedPasswordBase64)
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
