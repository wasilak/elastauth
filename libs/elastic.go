package libs

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

var tracerElastic = otel.Tracer("elastic")

// `var client *http.Client` is declaring a variable named `client` of type `*http.Client`. The `*`
// before `http.Client` indicates that `client` is a pointer to an instance of the `http.Client`
// struct. This variable is used to make HTTP requests to an Elasticsearch server.
var client *http.Client

// newHTTPClient creates a new HTTP client with optional TLS verification skipping.
// If insecure_skip_verify is enabled, it configures the client to skip certificate verification.
func newHTTPClient() *http.Client {
	httpClient := &http.Client{}
	if viper.GetBool("insecure_skip_verify") {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		httpClient.Transport = transport
	}
	return httpClient
}

// ElasticsearchConnectionDetails holds the connection configuration for an Elasticsearch cluster.
// It includes the cluster URL, username, and password required for authentication.
type ElasticsearchConnectionDetails struct {
	URL      string
	Username string
	Password string
}

// ElasticsearchUserMetadata contains additional metadata about an Elasticsearch user,
// particularly the groups they belong to for access control purposes.
type ElasticsearchUserMetadata struct {
	Groups []string `json:"groups"`
}

// ElasticsearchUser represents a user object in Elasticsearch with authentication credentials,
// metadata, and role assignments for access control.
type ElasticsearchUser struct {
	Enabled  bool                      `json:"enabled"`
	Email    string                    `json:"email"`
	Password string                    `json:"password"`
	Metadata ElasticsearchUserMetadata `json:"metadata"`
	FullName string                    `json:"full_name"`
	Roles    []string                  `json:"roles"`
}

// `var elasticsearchConnectionDetails ElasticsearchConnectionDetails` is declaring a variable named
// `elasticsearchConnectionDetails` of type `ElasticsearchConnectionDetails`. This variable is used to
// store the URL, username, and password information required to connect to an Elasticsearch instance.
// It is initialized to an empty `ElasticsearchConnectionDetails` struct when the program starts.
var elasticsearchConnectionDetails ElasticsearchConnectionDetails

// The function initializes an Elasticsearch client with connection details and sends a GET request to
// the Elasticsearch URL with basic authentication.
func initElasticClient(ctx context.Context, url, user, pass string) error {
	_, span := tracerElastic.Start(ctx, "initElasticClient")
	defer span.End()

	client = newHTTPClient()

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

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return fmt.Errorf("failed to decode Elasticsearch response: %w", err)
	}

	slog.DebugContext(ctx, "Request response", slog.Any("body", SanitizeForLogging(body)))

	return nil
}

// UpsertUser creates or updates a user in Elasticsearch with the provided credentials and configuration.
// It sends the user data to the Elasticsearch security API using basic authentication.
// If the user already exists, their data will be updated; otherwise, a new user is created.
func UpsertUser(ctx context.Context, username string, elasticsearchUser ElasticsearchUser) error {
	_, span := tracerElastic.Start(ctx, "UpsertUser")
	defer span.End()

	client = newHTTPClient()

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

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return fmt.Errorf("failed to decode Elasticsearch response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with status %d: %+v", resp.StatusCode, body)
	}

	slog.DebugContext(ctx, "Request response", slog.Any("body", SanitizeForLogging(body)))

	return nil
}
