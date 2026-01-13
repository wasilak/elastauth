# Requirements Document

## Introduction

Transform elastauth from an Authelia-specific authentication proxy into a pluggable authentication system that can work with any authentication provider. The system will maintain its core purpose as a stateless bridge between authentication systems and Elasticsearch/Kibana, while adding support for multiple authentication providers through a common interface.

## Glossary

- **Auth_Provider**: A pluggable component that implements the AuthProvider interface to extract user information from different authentication systems
- **Elastauth**: The main application that acts as a stateless converter between auth providers and Elasticsearch/Kibana user management
- **User_Info**: A standardized structure containing username, email, groups, and full name extracted from any auth provider
- **Auth_Request**: The incoming HTTP request containing authentication information in various formats (headers, tokens, cookies, etc.) depending on the provider
- **Provider_Config**: Configuration specific to each authentication provider type
- **Elasticsearch_User**: A local Elasticsearch user account with managed credentials and role mappings that works with both Elasticsearch and Kibana

## Requirements

### Requirement 1: Provider Interface Architecture

**User Story:** As a system architect, I want a common interface for all authentication providers, so that elastauth can work with any authentication system without code changes.

#### Acceptance Criteria

1. THE Auth_Provider interface SHALL define a GetUser method that accepts an Auth_Request and returns User_Info
2. THE Auth_Provider interface SHALL define a Type method that returns the provider type as a string
3. WHEN a provider is called, THE system SHALL return standardized User_Info regardless of the underlying auth system
4. THE User_Info struct SHALL contain username, email, groups, and full name fields
5. THE Auth_Request SHALL provide access to HTTP headers, query parameters, request body, and cookies to support different authentication mechanisms

### Requirement 2: Authelia Provider Implementation

**User Story:** As an existing elastauth user, I want the Authelia provider to work exactly like the current system, so that my deployment continues working without changes.

#### Acceptance Criteria

1. THE Authelia_Provider SHALL extract username from the Remote-User header
2. THE Authelia_Provider SHALL extract groups from the Remote-Groups header  
3. THE Authelia_Provider SHALL extract email from the Remote-Email header
4. THE Authelia_Provider SHALL extract full name from the Remote-Name header
5. WHEN headers are missing, THE Authelia_Provider SHALL return appropriate errors
6. THE Authelia_Provider SHALL maintain backward compatibility with existing configurations

### Requirement 3: Generic OAuth2/OIDC Provider Implementation

**User Story:** As a system administrator, I want to use any OAuth2/OIDC-compliant provider, so that I can integrate with existing identity systems like Casdoor, Keycloak, Authentik, Auth0, Azure AD, Pocket-ID, Ory Hydra, or Authelia (OIDC mode).

#### Acceptance Criteria

1. THE OIDC_Provider SHALL support standard OAuth2/OIDC authentication flows including authorization code flow
2. THE OIDC_Provider SHALL verify JWT tokens using the configured OIDC issuer and JWKS endpoint
3. THE OIDC_Provider SHALL extract user claims from validated JWT tokens or userinfo endpoint
4. THE OIDC_Provider SHALL support configurable claim mappings for username, email, groups, and full name
5. THE OIDC_Provider SHALL support standard OIDC discovery for automatic endpoint configuration
6. THE OIDC_Provider SHALL support manual endpoint configuration for non-discovery providers
7. WHEN token verification fails, THE OIDC_Provider SHALL return authentication errors with appropriate HTTP status codes
8. THE OIDC_Provider SHALL support both Bearer token authentication and cookie-based sessions
9. THE OIDC_Provider SHALL handle OAuth2 error responses according to RFC 6749 standards
10. THE OIDC_Provider SHALL support PKCE (Proof Key for Code Exchange) for enhanced security

### Requirement 4: OAuth2/OIDC Configuration Flexibility

**User Story:** As a system administrator, I want flexible OAuth2/OIDC configuration options, so that I can integrate with any OAuth2/OIDC provider regardless of their specific implementation details.

#### Acceptance Criteria

1. THE system SHALL support both OIDC discovery and manual endpoint configuration
2. THE system SHALL support configurable OAuth2 scopes for different provider requirements
3. THE system SHALL support both client_secret_basic and client_secret_post authentication methods
4. THE system SHALL support configurable token validation methods (JWKS, userinfo endpoint, or both)
5. THE system SHALL support configurable claim sources (ID token, access token, or userinfo endpoint)
6. THE system SHALL support custom HTTP headers for provider-specific requirements
7. THE system SHALL validate OAuth2/OIDC configuration at startup and provide clear error messages
8. THE system SHALL support both public and confidential OAuth2 client types

### Requirement 5: Single Provider Configuration System

**User Story:** As a system administrator, I want to configure exactly one authentication provider, so that the system has a clear and simple authentication path without complexity.

#### Acceptance Criteria

1. THE system SHALL read the auth_provider configuration to determine which single provider to use
2. THE system SHALL load only the configuration section for the selected provider
3. WHEN no auth_provider is specified, THE system SHALL default to "authelia" provider for backward compatibility
4. WHEN an invalid provider is specified, THE system SHALL return a configuration error and fail to start
5. THE system SHALL validate that exactly one provider is configured and active
6. WHEN multiple providers are configured, THE system SHALL return a configuration error and fail to start
7. THE system SHALL support environment variable overrides for sensitive configuration values

### Requirement 6: Provider Factory and Registration

**User Story:** As a developer, I want the selected provider to be automatically instantiated, so that the system starts with exactly one active provider.

#### Acceptance Criteria

1. THE Provider_Factory SHALL instantiate exactly one provider based on configuration
2. THE system SHALL register all available provider types at startup for selection
3. WHEN the selected provider fails to initialize, THE system SHALL log the error and exit gracefully
4. THE Provider_Factory SHALL validate provider configuration before instantiation
5. THE system SHALL prevent runtime provider switching to maintain simplicity and predictability

### Requirement 7: Backward Compatibility

**User Story:** As an existing elastauth user, I want my current configuration to continue working, so that I can upgrade without breaking my deployment.

#### Acceptance Criteria

1. WHEN no auth_provider is specified, THE system SHALL default to "authelia" provider
2. THE existing header configuration SHALL continue to work with the Authelia provider
3. THE existing group mappings and default roles SHALL work with all providers
4. THE existing cache behavior SHALL remain unchanged across all providers
5. THE existing API endpoints SHALL maintain the same response format

### Requirement 8: Error Handling and Logging

**User Story:** As a system administrator, I want clear error messages and logging, so that I can troubleshoot authentication issues effectively.

#### Acceptance Criteria

1. WHEN a provider fails to authenticate, THE system SHALL log the specific error with provider context
2. THE system SHALL return appropriate HTTP status codes for different error types
3. WHEN configuration is invalid, THE system SHALL provide clear error messages indicating the problem
4. THE system SHALL log the selected provider type and initialization status at startup
5. THE system SHALL sanitize sensitive information from logs while maintaining debugging capability

### Requirement 9: Stateless Operation with Cachego Multi-Provider Caching

**User Story:** As a platform engineer, I want elastauth to remain stateless for authentication decisions while using the cachego library for flexible caching, so that I can deploy multiple instances with various caching backends.

#### Acceptance Criteria

1. THE system SHALL NOT store any persistent authentication state between requests
2. THE system SHALL delegate all authentication decisions to the configured provider
3. THE system SHALL use the cachego library for all caching operations
4. THE system SHALL support cachego's memory, redis, and file cache providers
5. THE system SHALL support exactly zero or one cache type configured at any time
6. WHEN multiple cache types are configured, THE system SHALL return a configuration error and fail to start
7. WHEN caching is enabled, THE system SHALL store encrypted temporary passwords using cachego interface
8. WHEN caching is disabled, THE system SHALL generate new credentials and call Elasticsearch API on every request
9. FOR horizontal scaling with multiple instances, THE system SHALL require identical configuration including encryption keys
10. FOR horizontal scaling with multiple instances using redis cache, THE system SHALL share the Redis instance across all elastauth instances
11. FOR horizontal scaling with memory or file cache, THE system SHALL only support single instance deployment
12. THE system SHALL validate cache configuration compatibility with deployment mode at startup

### Requirement 10: Security and Validation

**User Story:** As a platform engineer, I want to configure multiple Elasticsearch endpoints for high availability, so that elastauth can failover between Elasticsearch instances without using version-specific SDKs.

#### Acceptance Criteria

1. THE system SHALL support a list of Elasticsearch endpoints for high availability
2. THE system SHALL use HTTP calls directly instead of Elasticsearch SDKs to avoid version coupling
3. WHEN the primary Elasticsearch endpoint fails, THE system SHALL attempt the next endpoint in the list
4. THE system SHALL validate that all configured Elasticsearch endpoints use the same authentication credentials
5. THE system SHALL log endpoint failover attempts and successes for monitoring
6. THE system SHALL support both single endpoint and multiple endpoint configurations
7. WHEN all Elasticsearch endpoints fail, THE system SHALL return appropriate error responses

### Requirement 12: API Documentation and Standards

**User Story:** As a developer integrating with elastauth, I want comprehensive API documentation with Swagger/OpenAPI specification and consistent JSON responses, so that I can understand and integrate with the API effectively.

#### Acceptance Criteria

1. THE system SHALL provide Swagger/OpenAPI specification for all API endpoints
2. THE system SHALL serve interactive API documentation at a dedicated endpoint
3. THE API SHALL follow RESTful conventions and HTTP status code standards
4. THE system SHALL return JSON responses for both success and error cases, never just HTTP status codes
5. THE system SHALL provide clear error response formats with consistent structure including error message and code
6. THE system SHALL document all request and response schemas in the OpenAPI specification
7. THE system SHALL include example requests and responses for both success and error scenarios in the API documentation
8. THE system SHALL validate API requests against the OpenAPI schema where possible

### Requirement 13: Configuration Visibility

**User Story:** As a system administrator, I want to view the current system configuration including auth provider and caching settings, so that I can verify the system is configured correctly without exposing sensitive credentials.

#### Acceptance Criteria

1. THE config endpoint SHALL return the currently active authentication provider type
2. THE config endpoint SHALL return the currently active cache configuration type
3. THE config endpoint SHALL return Elasticsearch endpoint configuration with masked credentials
4. THE config endpoint SHALL mask all sensitive configuration values (passwords, secrets, tokens)
5. THE config endpoint SHALL continue to return default roles and group mappings as it currently does
6. THE config endpoint SHALL return provider-specific configuration with sensitive values masked
7. THE config endpoint SHALL indicate which configuration values are masked for security

### Requirement 14: Kubernetes and Container Readiness

**User Story:** As a platform engineer, I want elastauth to be fully Kubernetes-ready with proper configuration management, so that I can deploy it in containerized environments with standard Kubernetes practices.

#### Acceptance Criteria

1. THE system SHALL support configuration via environment variables for all settings
2. THE system SHALL support configuration via mounted ConfigMaps and Secrets
3. THE system SHALL provide health check endpoints suitable for Kubernetes liveness and readiness probes
4. THE system SHALL log to stdout/stderr for proper container log collection
5. THE system SHALL handle SIGTERM gracefully for proper Kubernetes pod termination
6. THE system SHALL support configuration precedence: environment variables > config files > defaults
7. THE system SHALL validate that all required configuration is available at startup
8. THE system SHALL exit with appropriate error codes when configuration is invalid or missing

### Requirement 15: Comprehensive Documentation

**User Story:** As a user, developer, or system administrator, I want comprehensive documentation organized by topic, so that I can understand, configure, and operate elastauth effectively.

#### Acceptance Criteria

1. THE system SHALL provide a dedicated docs/ folder with structured documentation
2. THE main README SHALL link to specific documentation topics rather than containing all details
3. THE documentation SHALL include a concepts overview explaining elastauth's role and architecture
4. THE documentation SHALL include brief introductions to third-party components (Elasticsearch, auth systems, caching) with external links
5. THE documentation SHALL include a horizontal scaling guide with configuration requirements and constraints
6. THE documentation SHALL include an auth providers section with a dedicated page for each supported provider
7. THE documentation SHALL include a cache providers section with a dedicated page for each supported cache type
8. THE documentation SHALL use cross-linked Markdown files for easy navigation
9. THE documentation SHALL include configuration examples for each provider and cache type
10. THE documentation SHALL include troubleshooting guides for common issues

### Requirement 16: Security and Validation

**User Story:** As a security engineer, I want all user input to be validated, so that the system remains secure against malicious input.

#### Acceptance Criteria

1. THE system SHALL validate all user information returned by providers
2. THE system SHALL sanitize user input before processing or logging
3. WHEN invalid user data is received, THE system SHALL reject the request with appropriate errors
4. THE system SHALL validate provider configuration for security best practices
5. THE system SHALL encrypt cached credentials using the existing encryption mechanism

### Requirement 17: Documentation Migration to Starlight

**User Story:** As a user, developer, or system administrator, I want modern, searchable, and well-organized documentation hosted on GitHub Pages, so that I can easily find information and navigate between topics.

#### Acceptance Criteria

1. THE system SHALL migrate all existing documentation from the docs/ folder to Starlight framework
2. THE documentation SHALL be built using Astro with the Starlight template using default styling without custom CSS
3. THE documentation SHALL include full-text search functionality using Pagefind (built-in with Starlight)
4. THE documentation SHALL support light/dark/system theme modes out of the box
5. THE documentation SHALL maintain all existing content including concepts, provider guides, cache documentation, and troubleshooting
6. THE documentation SHALL use Starlight's sidebar navigation with proper categorization and auto-generation where appropriate
7. THE documentation SHALL be deployed to GitHub Pages using the provided workflow configuration with Node.js 24 and Task runner
8. THE documentation SHALL include proper SEO optimization and meta tags provided by Starlight
9. THE documentation SHALL remain English-only without internationalization configuration
10. THE documentation SHALL include edit links pointing to the GitHub repository for community contributions
11. THE documentation SHALL support Mermaid diagrams using rehype-mermaid plugin for architecture and flow diagrams
12. THE documentation SHALL include YAML configuration examples but no code snippets except for deployment configurations
13. THE documentation SHALL maintain cross-references and internal linking between documentation pages