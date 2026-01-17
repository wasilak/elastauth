package libs

import (
	"bytes"
	"context"
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

// ElasticsearchConnectionDetails holds the connection configuration for an Elasticsearch cluster.
// It includes the cluster URL, username, and password required for authentication.
type ElasticsearchConnectionDetails struct {
	Hosts    []string
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
func initElasticClient(ctx context.Context, hosts []string, user, pass string) error {
	_, span := tracerElastic.Start(ctx, "initElasticClient")
	defer span.End()

	if len(hosts) == 0 {
		return fmt.Errorf("no Elasticsearch hosts provided")
	}

	client = &http.Client{}

	elasticsearchConnectionDetails = ElasticsearchConnectionDetails{
		Hosts:    hosts,
		Username: user,
		Password: pass,
	}

	// Validate all endpoints have the same credentials and are accessible
	var lastErr error
	successfulHosts := 0

	for i, host := range hosts {
		slog.DebugContext(ctx, "Testing Elasticsearch endpoint", slog.String("host", host), slog.Int("index", i))
		
		req, err := http.NewRequest("GET", host, nil)
		if err != nil {
			slog.WarnContext(ctx, "Failed to create request for Elasticsearch endpoint", 
				slog.String("host", host), slog.String("error", err.Error()))
			lastErr = err
			continue
		}

		req.Header.Add("Authorization", "Basic "+basicAuth(user, pass))
		resp, err := client.Do(req)
		if err != nil {
			slog.WarnContext(ctx, "Failed to connect to Elasticsearch endpoint", 
				slog.String("host", host), slog.String("error", err.Error()))
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			slog.WarnContext(ctx, "Elasticsearch endpoint returned non-200 status", 
				slog.String("host", host), slog.Int("status", resp.StatusCode))
			lastErr = fmt.Errorf("endpoint %s returned status %d", host, resp.StatusCode)
			continue
		}

		body := map[string]interface{}{}
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			slog.WarnContext(ctx, "Failed to decode Elasticsearch response", 
				slog.String("host", host), slog.String("error", err.Error()))
			lastErr = fmt.Errorf("failed to decode response from %s: %w", host, err)
			continue
		}

		slog.DebugContext(ctx, "Successfully connected to Elasticsearch endpoint", 
			slog.String("host", host), slog.Any("response", SanitizeForLogging(body)))
		successfulHosts++
	}

	if successfulHosts == 0 {
		return fmt.Errorf("failed to connect to any Elasticsearch endpoint, last error: %w", lastErr)
	}

	slog.InfoContext(ctx, "Elasticsearch client initialized", 
		slog.Int("total_hosts", len(hosts)), 
		slog.Int("successful_hosts", successfulHosts))

	return nil
}

// UpsertUser creates or updates a user in Elasticsearch with the provided credentials and configuration.
// It sends the user data to the Elasticsearch security API using basic authentication.
// If the user already exists, their data will be updated; otherwise, a new user is created.
func UpsertUser(ctx context.Context, username string, elasticsearchUser ElasticsearchUser) error {
	_, span := tracerElastic.Start(ctx, "UpsertUser")
	defer span.End()

	client = &http.Client{}

	var lastErr error
	
	// Try each Elasticsearch endpoint until one succeeds
	for i, host := range elasticsearchConnectionDetails.Hosts {
		slog.DebugContext(ctx, "Attempting to upsert user", 
			slog.String("username", username), 
			slog.String("host", host), 
			slog.Int("attempt", i+1))

		url := fmt.Sprintf("%s/_security/user/%s", host, username)

		jsonPayload, err := json.Marshal(elasticsearchUser)
		if err != nil {
			return fmt.Errorf("failed to marshal user data: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			slog.WarnContext(ctx, "Failed to create request for Elasticsearch endpoint", 
				slog.String("host", host), slog.String("error", err.Error()))
			lastErr = err
			continue
		}

		req.Header.Add("Authorization", "Basic "+basicAuth(elasticsearchConnectionDetails.Username, elasticsearchConnectionDetails.Password))
		req.Header.Add("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err != nil {
			slog.WarnContext(ctx, "Failed to connect to Elasticsearch endpoint for user upsert", 
				slog.String("host", host), slog.String("username", username), slog.String("error", err.Error()))
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		body := map[string]interface{}{}
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			slog.WarnContext(ctx, "Failed to decode Elasticsearch response", 
				slog.String("host", host), slog.String("username", username), slog.String("error", err.Error()))
			lastErr = fmt.Errorf("failed to decode response from %s: %w", host, err)
			continue
		}

		if resp.StatusCode != 200 {
			slog.WarnContext(ctx, "Elasticsearch endpoint returned error for user upsert", 
				slog.String("host", host), 
				slog.String("username", username), 
				slog.Int("status", resp.StatusCode),
				slog.Any("response", SanitizeForLogging(body)))
			lastErr = fmt.Errorf("request to %s failed with status %d: %+v", host, resp.StatusCode, body)
			continue
		}

		slog.DebugContext(ctx, "Successfully upserted user", 
			slog.String("username", username), 
			slog.String("host", host),
			slog.Any("response", SanitizeForLogging(body)))

		return nil
	}

	// If we get here, all endpoints failed
	return fmt.Errorf("failed to upsert user %s to any Elasticsearch endpoint, last error: %w", username, lastErr)
}

// GetElasticsearchHosts returns the list of Elasticsearch hosts from configuration
// Supports both new multi-endpoint format and legacy single host format
func GetElasticsearchHosts() []string {
	// Try new format first
	hosts := viper.GetStringSlice("elasticsearch.hosts")
	if len(hosts) > 0 {
		return hosts
	}

	// Fall back to legacy format
	legacyHost := viper.GetString("elasticsearch_host")
	if legacyHost != "" {
		return []string{legacyHost}
	}

	return []string{}
}

// GetElasticsearchUsername returns the Elasticsearch username from configuration
// Supports both new and legacy configuration formats
func GetElasticsearchUsername() string {
	// Try new format first
	username := viper.GetString("elasticsearch.username")
	if username != "" {
		return username
	}

	// Fall back to legacy format
	return viper.GetString("elasticsearch_username")
}

// GetElasticsearchPassword returns the Elasticsearch password from configuration
// Supports both new and legacy configuration formats
func GetElasticsearchPassword() string {
	// Try new format first
	password := viper.GetString("elasticsearch.password")
	if password != "" {
		return password
	}

	// Fall back to legacy format
	return viper.GetString("elasticsearch_password")
}

// GetElasticsearchDryRun returns the Elasticsearch dry run setting from configuration
// Supports both new and legacy configuration formats
func GetElasticsearchDryRun() bool {
	// Try new format first
	if viper.IsSet("elasticsearch.dry_run") {
		return viper.GetBool("elasticsearch.dry_run")
	}

	// Fall back to legacy format
	return viper.GetBool("elasticsearch_dry_run")
}
