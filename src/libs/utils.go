package libs

import (
	"encoding/base64"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/viper"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func getMapKeys(itemsMap map[string][]string) []string {
	keys := []string{}

	for k, _ := range itemsMap {
		keys = append(keys, k)
	}

	return keys
}

func GenerateTemporaryUserPassword() string {
	// Generate a password that is 32 characters long with 6 digits, 6 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	res, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func GetUserRoles(userGroups []string) []string {
	roles := []string{}
	if len(viper.GetStringMapStringSlice("group_mappings")) > 0 {
		for _, group := range userGroups {
			if contains(getMapKeys(viper.GetStringMapStringSlice("group_mappings")), group) {
				for _, mappings := range viper.GetStringMapStringSlice("group_mappings") {
					for _, mappingName := range mappings {
						roles = append(roles, mappingName)
					}
				}
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

func EncryptPassword(pass, encryptionKey string) string {
	var (
		password = []byte(encryptionKey)
		data     = []byte(pass)
	)

	ciphertext, err := Encrypt(password, data)
	if err != nil {
		log.Fatal(err)
	}

	return string(ciphertext)
}

func DecryptPassword(encryptedString string, encryptionKey string) string {
	var (
		password = []byte(encryptionKey)
	)

	plaintext, err := Decrypt(password, []byte(encryptedString))
	if err != nil {
		log.Fatal(err)
	}

	return string(plaintext)
}
