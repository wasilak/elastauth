package libs

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/sethvargo/go-password/password"
	"github.com/spf13/viper"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(strings.ToLower(v), strings.ToLower(str)) {
			return true
		}
	}
	return false
}

func getMapKeys(itemsMap map[string][]string) []string {
	keys := []string{}

	for k := range itemsMap {
		keys = append(keys, k)
	}

	return keys
}

func GenerateTemporaryUserPassword() (string, error) {
	// Generate a password that is 32 characters long with 6 digits, 6 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	res, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		return "", err
	}
	return res, nil
}

func GetUserRoles(userGroups []string) []string {
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

func basicAuth(username, pass string) string {
	auth := username + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GenerateKey() (string, error) {
	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil //encode key in bytes to string for saving

}
