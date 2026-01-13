# Implementation Plan: Pluggable Authentication Providers

## Overview

Transform elastauth from an Authelia-specific authentication proxy into a pluggable authentication system. The implementation follows a phase-gate approach, building incrementally on the existing working system while maintaining backward compatibility. The system will support two authentication providers: Authelia (header-based) for backward compatibility, and a generic OAuth2/OIDC provider that works with any OAuth2/OIDC-compliant system including Casdoor, Keycloak, Authentik, Auth0, Azure AD, Pocket-ID, Ory Hydra, and others.

## Tasks

- [x] 1. Phase 1: Provider Interface Foundation
- [x] 1.1 Create core provider interfaces and types
  - Create `provider/` package with `AuthProvider` interface
  - Define `AuthRequest`, `UserInfo`, and `Factory` types
  - Implement provider registration system
  - _Requirements: 1.1, 1.2, 1.4, 6.1, 6.2_

- [ ]* 1.2 Write property test for provider interface compliance
  - **Property 1: Provider Interface Compliance**
  - **Validates: Requirements 1.1, 1.2, 1.4**

- [x] 1.3 Implement Authelia provider with backward compatibility
  - Create `provider/authelia/` package
  - Implement `AutheliaProvider` struct and methods
  - Extract user info from existing headers (Remote-User, Remote-Groups, etc.)
  - Maintain exact compatibility with current behavior
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 7.2_

- [ ]* 1.4 Write property test for Authelia header extraction
  - **Property 3: Authelia Header Extraction**
  - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

- [ ]* 1.5 Write property test for provider error handling
  - **Property 4: Provider Error Handling**
  - **Validates: Requirements 2.5**

- [x] 1.6 Integrate provider system with existing MainRoute
  - Modify `libs/routes.go` to use provider factory
  - Replace direct header extraction with provider calls
  - Ensure cache and Elasticsearch integration remains unchanged
  - _Requirements: 1.3, 7.4, 7.5_

- [ ]* 1.7 Write property test for standardized user information
  - **Property 2: Standardized User Information**
  - **Validates: Requirements 1.3**

- [x] 1.8 Phase 1 Checkpoint - Verify backward compatibility
  - Ensure all tests pass, ask the user if questions arise
  - Verify existing Authelia functionality works unchanged
  - Test with real Authelia headers and configuration

- [ ] 2. Phase 2: Configuration System Enhancement
- [x] 2.1 Enhance configuration structure for multiple providers and cachego
  - Update `libs/config.go` to support `auth_provider` selection
  - Add provider-specific configuration sections
  - Integrate cachego library for cache management
  - Add cachego configuration with backward compatibility
  - Implement configuration validation for single provider selection
  - _Requirements: 5.1, 5.2, 5.4, 5.5, 9.3, 9.4, 9.5_

- [ ]* 2.2 Write property test for single provider configuration
  - **Property 5: Single Provider Configuration**
  - **Validates: Requirements 5.1, 5.5, 6.1**

- [x] 2.3 Implement cachego integration and migration
  - Replace existing cache interface with cachego library
  - Add support for memory, redis, and file cache providers
  - Implement backward compatibility with existing cache configuration
  - Add configuration mapping from legacy to cachego format
  - _Requirements: 9.3, 9.4, 9.5, 9.7, 9.10, 9.11_

- [ ]* 2.4 Write property test for cachego cache providers
  - **Property 7: Cache Configuration Validation**
  - **Validates: Requirements 9.3, 9.4, 9.5**

- [x] 2.5 Implement default provider selection and validation
  - Default to "oidc" when no provider specified
  - Validate exactly one provider is configured
  - Return configuration errors for invalid setups
  - _Requirements: 5.3, 5.6, 6.3, 6.4_

- [x] 2.6 Add environment variable support for provider configuration
  - Extend Viper configuration for provider-specific env vars
  - Implement configuration precedence rules
  - Support sensitive value overrides
  - _Requirements: 5.7, 14.1, 14.6_

- [ ]* 2.7 Write property test for configuration precedence
  - **Property 6: Configuration Precedence**
  - **Validates: Requirements 5.7, 14.6**

- [x] 2.8 Update configuration endpoint to show provider and cache information
  - Enhance `ConfigRoute` to return active provider type
  - Add cachego cache status and configuration
  - Add provider-specific configuration with masked sensitive values
  - Include cache configuration information
  - _Requirements: 13.1, 13.2, 13.4, 13.6_

- [ ]* 2.9 Write property test for configuration masking
  - **Property 12: Configuration Masking**
  - **Validates: Requirements 13.4, 13.6**

- [x] 2.10 Phase 2 Checkpoint - Configuration system works
  - Ensure all tests pass, ask the user if questions arise
  - Verify provider selection via configuration
  - Test environment variable overrides
  - Verify cachego integration works with all supported providers

- [x] 3. Phase 3: Generic OAuth2/OIDC Provider Implementation
- [x] 3.1 Research OAuth2/OIDC standards and Go libraries
  - Investigate golang.org/x/oauth2 and coreos/go-oidc libraries
  - Design generic OAuth2/OIDC integration approach
  - Plan support for multiple authentication flows and token validation methods

- [x] 3.2 Implement OAuth2/OIDC provider structure
  - Create `provider/oidc/` package
  - Implement `OIDCProvider` struct with comprehensive configuration
  - Add support for both OIDC discovery and manual endpoint configuration
  - Implement OAuth2 client setup with PKCE support
  - _Requirements: 3.1, 3.2, 3.5, 3.6, 4.1, 4.2, 4.7_

- [x] 3.3 Implement token validation and user information extraction
  - Implement JWT token verification using JWKS
  - Implement userinfo endpoint validation as alternative
  - Support both Bearer token and cookie-based authentication
  - Add configurable claim mapping with support for nested claims
  - Handle OAuth2 error responses according to RFC standards
  - _Requirements: 3.2, 3.3, 3.7, 3.8, 4.4, 4.5_

- [ ]* 3.4 Write property test for OAuth2/OIDC provider functionality
  - Test token validation, claim extraction, and error handling
  - Test configurable claim mappings with various claim structures
  - **Validates: Requirements 3.2, 3.3, 3.4, 3.7, 3.9**

- [x] 3.5 Implement OAuth2/OIDC configuration validation
  - Add comprehensive configuration validation for OAuth2/OIDC settings
  - Support both discovery and manual endpoint configuration
  - Validate client authentication methods and token validation options
  - Add support for custom headers and provider-specific requirements
  - _Requirements: 4.1, 4.3, 4.6, 4.7, 4.8_

- [ ]* 3.6 Write property test for OAuth2/OIDC configuration validation
  - Test configuration validation for various OAuth2/OIDC scenarios
  - **Validates: Requirements 4.1, 4.7**

- [x] 3.7 Register OAuth2/OIDC provider in factory
  - Add OIDC provider to provider registration
  - Update configuration validation for OIDC options
  - Test provider selection and instantiation with various configurations
  - _Requirements: 6.1, 6.2, 6.4_

- [x] 3.8 Phase 3 Checkpoint - OAuth2/OIDC provider works
  - Ensure all tests pass, ask the user if questions arise
  - Test OAuth2/OIDC provider with test JWT tokens from different issuers
  - Verify integration with existing elastauth flow
  - Test with example configurations for Casdoor, Keycloak, and Authentik

- [-] 4. Phase 4: Enhanced API and Documentation
- [x] 4.1 Implement Swagger/OpenAPI specification
  - Add OpenAPI specification for all API endpoints
  - Include request and response schemas
  - Add example requests and responses for success and error cases
  - _Requirements: 12.1, 12.6, 12.7_

- [x] 4.2 Serve interactive API documentation
  - Add endpoint to serve Swagger UI
  - Ensure API documentation is accessible and interactive
  - Validate API requests against OpenAPI schema where possible
  - _Requirements: 12.2, 12.8_

- [x] 4.3 Enhance API response consistency
  - Ensure all endpoints return JSON responses (never just status codes)
  - Implement consistent error response structure
  - Follow RESTful conventions and proper HTTP status codes
  - _Requirements: 12.3, 12.4, 12.5_

- [ ]* 4.4 Write property test for JSON response consistency
  - **Property 11: JSON Response Consistency**
  - **Validates: Requirements 12.4, 12.5**

- [x] 4.5 Create comprehensive documentation structure
  - Create `docs/` folder with structured documentation
  - Write concepts overview and architecture documentation
  - Create OAuth2/OIDC provider documentation with examples for popular providers
  - Add configuration examples and troubleshooting guides
  - _Requirements: 15.1, 15.3, 15.6, 15.9, 15.10_

- [x] 4.6 Update main README with documentation links
  - Simplify README to link to specific documentation topics
  - Add cross-links between documentation files
  - Include horizontal scaling guide
  - _Requirements: 15.2, 15.5, 15.8_

- [x] 4.7 Phase 4 Checkpoint - API and documentation complete
  - Ensure all tests pass, ask the user if questions arise
  - Verify Swagger UI works and shows all endpoints
  - Review documentation completeness

- [-] 5. Phase 5: Multi-Elasticsearch and Production Readiness
- [-] 5.1 Implement multi-endpoint Elasticsearch support
  - Enhance `libs/elastic.go` to support multiple endpoints
  - Implement failover logic for Elasticsearch connectivity
  - Add endpoint validation and logging
  - _Requirements: 10.1, 10.3, 10.4, 10.5, 10.6, 10.7_

- [ ]* 5.2 Write property test for Elasticsearch failover
  - **Property 10: Elasticsearch Failover**
  - **Validates: Requirements 10.3, 10.7**

- [ ] 5.3 Enhance cachego configuration validation and multi-provider support
  - Implement validation for exactly zero or one cache type
  - Add cache configuration compatibility checks with cachego
  - Support cache-disabled scenarios
  - Test memory, redis, and file cache providers
  - Validate horizontal scaling constraints per cache type
  - _Requirements: 9.3, 9.4, 9.5, 9.10, 9.11_

- [ ]* 5.4 Write property test for cachego configuration validation
  - **Property 7: Cache Configuration Validation**
  - **Validates: Requirements 9.3, 9.4, 9.5**

- [ ] 5.5 Implement Kubernetes readiness features
  - Enhance health endpoints for liveness and readiness probes
  - Implement graceful SIGTERM handling
  - Ensure proper stdout/stderr logging
  - Validate configuration at startup with proper exit codes
  - _Requirements: 14.3, 14.4, 14.5, 14.7, 14.8_

- [ ]* 5.6 Write property test for input validation and sanitization
  - **Property 13: Input Validation and Sanitization**
  - **Validates: Requirements 16.1, 16.2, 16.3**

- [ ]* 5.7 Write property test for credential encryption
  - **Property 14: Credential Encryption**
  - **Validates: Requirements 16.5**

- [ ] 5.8 Final integration testing and validation
  - Test OAuth2/OIDC provider with real authentication systems (Casdoor, Keycloak, Authentik)
  - Verify cache behavior across all providers
  - Test Elasticsearch failover scenarios
  - Validate Kubernetes deployment readiness
  - _Requirements: 9.1, 9.2, 9.5, 9.6_

- [ ]* 5.9 Write property test for stateless authentication
  - **Property 8: Stateless Authentication**
  - **Validates: Requirements 9.1, 9.2**

- [ ]* 5.10 Write property test for cache behavior consistency
  - **Property 9: Cache Behavior Consistency**
  - **Validates: Requirements 9.5, 9.6**

- [ ] 5.11 Phase 5 Checkpoint - Production readiness complete
  - Ensure all tests pass, ask the user if questions arise
  - Verify system works in production-like environment
  - Test horizontal scaling scenarios

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at each phase
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Phase gates must pass before proceeding to next phase