package libs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"log/slog"

	"go.opentelemetry.io/otel"
)

var tracerElastic = otel.Tracer("elastic")

// `var client *http.Client` is declaring a variable named `client` of type `*http.Client`. The `*`
// before `http.Client` indicates that `client` is a pointer to an instance of the `http.Client`
// struct. This variable is used to make HTTP requests to an Elasticsearch server.
var client *http.Client

// The type `ElasticsearchConnectionDetails` contains URL, username, and password information for
// connecting to Elasticsearch.
// @property {string} URL - The URL property is a string that represents the endpoint of the
// Elasticsearch cluster that the application will connect to. It typically includes the protocol (http
// or https), the hostname or IP address of the Elasticsearch server, and the port number.
// @property {string} Username - The `Username` property is a string that represents the username used
// to authenticate the connection to an Elasticsearch instance.
// @property {string} Password - The `Password` property is a string that stores the password required
// to authenticate and establish a connection to an Elasticsearch instance. This property is typically
// used in conjunction with the `Username` property to provide secure access to the Elasticsearch
// cluster.
type ElasticsearchConnectionDetails struct {
	URL      string
	Username string
	Password string
}

// The type `ElasticsearchUserMetadata` contains a field `Groups` which is a slice of strings
// representing user groups.
// @property {[]string} Groups - The `Groups` property is a slice of strings that represents the groups
// that a user belongs to in Elasticsearch. This metadata can be used to control access to specific
// resources or features within Elasticsearch based on a user's group membership.
type ElasticsearchUserMetadata struct {
	Groups []string `json:"groups"`
}

// The ElasticsearchUser type represents a user in Elasticsearch with properties such as email,
// password, metadata, full name, and roles.
// @property {bool} Enabled - A boolean value indicating whether the Elasticsearch user is enabled or
// disabled.
// @property {string} Email - The email address of the Elasticsearch user.
// @property {string} Password - The "Password" property is a string that represents the password of an
// Elasticsearch user. It is used to authenticate the user when they try to access Elasticsearch
// resources. It is important to keep this property secure and encrypted to prevent unauthorized access
// to Elasticsearch data.
// @property {ElasticsearchUserMetadata} Metadata - Metadata is a property of the ElasticsearchUser
// struct that contains additional information about the user. It is of type ElasticsearchUserMetadata,
// which is likely another struct that contains specific metadata properties such as creation date,
// last login time, etc. The purpose of this property is to provide additional context and information
// about the
// @property {string} FullName - The FullName property is a string that represents the full name of an
// Elasticsearch user. It is one of the properties of the ElasticsearchUser struct.
// @property {[]string} Roles - Roles is a property of the ElasticsearchUser struct that represents the
// list of roles assigned to the user. Roles are used to define the level of access and permissions a
// user has within the Elasticsearch system. For example, a user with the "admin" role may have full
// access to all Elasticsearch features, while
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

	slog.DebugContext(ctx, "Request response", slog.Any("body", body))

	return nil
}

// The function UpsertUser sends a POST request to Elasticsearch to create or update a user with the
// given username and user details.
func UpsertUser(ctx context.Context, username string, elasticsearchUser ElasticsearchUser) error {
	_, span := tracerElastic.Start(ctx, "UpsertUser")
	defer span.End()

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

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed: %+v", body)
	}

	slog.DebugContext(ctx, "Request response", slog.Any("body", body))

	return nil
}
