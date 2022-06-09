package libs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

// var rdb *redis.Client
var ctx = context.Background()

var rdb *redis.Client

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
	// var encryptedPassword, decryptedPassword string
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

	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", viper.GetString("redis_host"), viper.GetString("redis_port")),
		DB:   viper.GetInt("redis_db"),
	})

	cacheKey := "elastauth-" + user

	cacheDuration, err := time.ParseDuration(viper.GetString("redis_expire_seconds"))
	if err != nil {
		log.Fatal(err)
	}

	encryptedPassword, err := rdb.Get(ctx, cacheKey).Result()
	// password, err := rdb.Get(ctx, cacheKey).Result()

	if err == redis.Nil {
		roles := GetUserRoles(userGroups)

		userEmail := c.Request().Header.Get("Remote-Email")
		userName := c.Request().Header.Get("Remote-Name")

		initElasticClient(
			viper.GetString("elasticsearch_host"),
			viper.GetString("elasticsearch_username"),
			viper.GetString("elasticsearch_password"),
		)

		password := GenerateTemporaryUserPassword()
		encryptedPassword = EncryptPassword(password, viper.GetString("secret_key"))

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

		rdb.Set(ctx, cacheKey, encryptedPassword, cacheDuration)
		// rdb.Set(ctx, cacheKey, password, 600*time.Second)
	} else if err != nil {
		panic(err)
	}

	itemCacheDuration, err := rdb.TTL(ctx, cacheKey).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(cacheDuration, itemCacheDuration)

	if viper.GetBool("extend_cache") && itemCacheDuration > 0 && itemCacheDuration < cacheDuration {
		log.Debug(fmt.Sprintf("User %s: extending cache TTL (from %s to %s)", user, itemCacheDuration, viper.GetString("redis_expire_seconds")))
		rdb.Expire(ctx, cacheKey, cacheDuration)
	}

	decryptedPassword := DecryptPassword(encryptedPassword, viper.GetString("secret_key"))

	c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, decryptedPassword))
	// c.Response().Header().Set(echo.HeaderAuthorization, "Basic "+basicAuth(user, password))

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
