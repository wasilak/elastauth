package libs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/gommon/log"
)

var client *http.Client

type ElasticsearchConnectionDetails struct {
	URL      string
	Username string
	Password string
}

type ElasticsearchUserMetadata struct {
	Groups []string `json:"groups"`
}

type ElasticsearchUser struct {
	Enabled  bool                      `json:"enabled"`
	Email    string                    `json:"email"`
	Password string                    `json:"password"`
	Metadata ElasticsearchUserMetadata `json:"metadata"`
	FullName string                    `json:"full_name"`
	Roles    []string                  `json:"roles"`
}

var elasticsearchConnectionDetails ElasticsearchConnectionDetails

func initElasticClient(url, user, pass string) error {
	client = &http.Client{}

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		URL:      url,
		Username: user,
		Password: pass,
	}

	req, err := http.NewRequest("GET", elasticsearchConnectionDetails.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(elasticsearchConnectionDetails.Username, elasticsearchConnectionDetails.Password))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body := map[string]interface{}{}

	json.NewDecoder(resp.Body).Decode(&body)

	log.Debug(body)

	return nil
}

func UpsertUser(username string, elasticsearchUser ElasticsearchUser) error {
	client = &http.Client{}

	url := fmt.Sprintf("%s/_security/user/%s", elasticsearchConnectionDetails.URL, username)

	jsonPayload, err := json.Marshal(elasticsearchUser)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(elasticsearchConnectionDetails.Username, elasticsearchConnectionDetails.Password))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body := map[string]interface{}{}

	json.NewDecoder(resp.Body).Decode(&body)

	log.Debug(body)

	return nil
}
