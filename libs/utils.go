package libs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"

	"github.com/sethvargo/go-password/password"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

var tracerUtils = otel.Tracer("utils")

// The function checks if a given string is present in a slice of strings, ignoring case sensitivity.
func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(strings.ToLower(v), strings.ToLower(str)) {
			return true
		}
	}
	return false
}

// The function returns an array of keys from a given map.
func getMapKeys(itemsMap map[string][]string) []string {
	keys := []string{}

	for k := range itemsMap {
		keys = append(keys, k)
	}

	return keys
}

// The function generates a temporary user password that is 32 characters long with a mix of digits,
// symbols, and upper/lower case letters, disallowing repeat characters.
func GenerateTemporaryUserPassword(ctx context.Context) (string, error) {
	_, span := tracerUtils.Start(ctx, "GenerateTemporaryUserPassword")
	defer span.End()

	// `res, err := password.Generate(32, 10, 0, false, false)` is generating a temporary user password
	// that is 32 characters long with a mix of digits, symbols, and upper/lower case letters, disallowing
	// repeat characters. It uses the `go-password` package to generate the password and returns the
	// generated password and any error that occurred during the generation process.
	res, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		return "", err
	}
	return res, nil
}

// The function retrieves user roles based on their group mappings or default roles if no mappings are
// found.
func GetUserRoles(ctx context.Context, userGroups []string) []string {
	_, span := tracerUtils.Start(ctx, "GetUserRoles")
	defer span.End()

	// This code block is retrieving user roles based on their group mappings or default roles if no
	// mappings are found.
	roles := []string{}
	if len(viper.GetStringMapStringSlice("group_mappings")) > 0 {
		for _, group := range userGroups {
			if contains(getMapKeys(viper.GetStringMapStringSlice("group_mappings")), group) {
				roles = append(roles, viper.GetStringMapStringSlice("group_mappings")[strings.ToLower(group)]...)
			}
		}
	}

	if len(roles) == 0 {
		return viper.GetStringSlice("default_roles")
	}

	return roles
}

// The function takes a username and password, combines them into a string, encodes the string using
// base64, and returns the encoded string.
func basicAuth(username, pass string) string {
	auth := username + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// The function generates a random 32 byte key for AES-256 encryption and returns it as a hexadecimal
// encoded string.
func GenerateKey(ctx context.Context) (string, error) {
	_, span := tracerUtils.Start(ctx, "GenerateKey")
	defer span.End()

	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil //encode key in bytes to string for saving

}

func GetAppName() string {
	appName := os.Getenv("OTEL_SERVICE_NAME")
	if appName == "" {
		appName = os.Getenv("APP_NAME")
		if appName == "" {
			appName = "elastauth"
		}
	}
	return appName
}
