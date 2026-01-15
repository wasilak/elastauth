# Implementation Tasks

## Phase 1: Forward Auth Mode (Header-Based)

### Task 1.1: Add Configuration Parameters

**Description**: Add `operation_mode` and `base_path` configuration parameters to the Config struct.

**Acceptance Criteria**:
- [ ] Add `operation_mode` field to Config struct with validator tag `required,oneof=forward-auth direct-auth`
- [ ] Add `base_path` field to Config struct with validator tag `required,startswith=/,excludes=//`
- [ ] Set default value for `base_path` to `/elastauth`
- [ ] Add environment variable support: `ELASTAUTH_OPERATION_MODE`, `ELASTAUTH_BASE_PATH`
- [ ] Add CLI flag support: `--operation-mode`, `--base-path`
- [ ] Update config.yml.example with new parameters

**Files to Modify**:
- `libs/config.go`
- `config.yml.example`

**Dependencies**: None

---

### Task 1.2: Refactor Configuration Structure

**Description**: Reorganize configuration into global and mode-specific sections.

**Acceptance Criteria**:
- [ ] Create `AutheliaConfig` struct with header field names
- [ ] Create `OIDCConfig` struct (placeholder for Phase 2)
- [ ] Move cache, logging, tracing, elasticsearch settings to global scope
- [ ] Add `Authelia` field to Config struct with validator tag `required_if=OperationMode forward-auth`
- [ ] Add `OIDC` field to Config struct with validator tag `required_if=OperationMode direct-auth`
- [ ] Update Viper unmarshaling to handle nested structures

**Files to Modify**:
- `libs/config.go`

**Dependencies**: Task 1.1

---

### Task 1.3: Implement Configuration Validation

**Description**: Add go-playground/validator for configuration validation.

**Acceptance Criteria**:
- [ ] Add `github.com/go-playground/validator/v10` dependency
- [ ] Create `RegisterCustomValidators()` function
- [ ] Implement `validateModeCompatibility` custom validator
- [ ] Implement `validateBasePath` custom validator
- [ ] Create `LoadConfig()` function that validates configuration
- [ ] Add validation for secret_key format (64 hex characters)
- [ ] Add validation for TLS certificate file existence
- [ ] Return descriptive error messages for validation failures

**Files to Create**:
- `libs/validation.go`

**Files to Modify**:
- `libs/config.go`
- `go.mod`

**Dependencies**: Task 1.2

---

### Task 1.4: Update Router with Base Path Support

**Description**: Refactor router to use configurable base_path for internal endpoints.

**Acceptance Criteria**:
- [ ] Create `Router` struct with `basePath` field
- [ ] Implement `NewRouter()` constructor
- [ ] Implement `ServeHTTP()` method with base path prefix checking
- [ ] Implement `handleInternalPath()` for routing internal endpoints
- [ ] Update health endpoint to be accessible at `{base_path}/health`
- [ ] Update ready endpoint to be accessible at `{base_path}/ready`
- [ ] Update config endpoint to be accessible at `{base_path}/config`
- [ ] Update metrics endpoint to be accessible at `{base_path}/metrics`
- [ ] Preserve backward compatibility for root path requests

**Files to Create**:
- `libs/router.go`

**Files to Modify**:
- `main.go`
- `libs/routes.go`

**Dependencies**: Task 1.1

---

### Task 1.5: Update Config Endpoint

**Description**: Update /config endpoint to show operation mode and provider information.

**Acceptance Criteria**:
- [ ] Add `operation_mode` to config endpoint response
- [ ] Add `base_path` to config endpoint response
- [ ] Add `auth_provider` to config endpoint response (based on mode)
- [ ] Add `proxy_enabled` to config endpoint response
- [ ] Ensure sensitive data (secrets, passwords) is not exposed
- [ ] Return JSON response with proper content type

**Files to Modify**:
- `libs/routes.go`

**Dependencies**: Task 1.4

---

### Task 1.6: Update Provider Factory

**Description**: Update provider factory to be mode-aware.

**Acceptance Criteria**:
- [ ] Create `ProviderFactory` struct with config field
- [ ] Implement `NewProviderFactory()` constructor
- [ ] Implement `GetProvider()` method that returns provider based on operation_mode
- [ ] For `forward-auth` mode, return Authelia provider
- [ ] For `direct-auth` mode, return error (not implemented yet)
- [ ] Add validation that provider matches operation mode

**Files to Create**:
- `provider/factory.go`

**Files to Modify**:
- `main.go`

**Dependencies**: Task 1.3

---

### Task 1.7: Update Authelia Provider

**Description**: Update Authelia provider to use new configuration structure.

**Acceptance Criteria**:
- [ ] Update `Provider` struct to use `AutheliaConfig`
- [ ] Update `NewProvider()` to accept `AutheliaConfig`
- [ ] Update header extraction to use config field names
- [ ] Ensure backward compatibility with existing behavior
- [ ] Add logging for header extraction at debug level

**Files to Modify**:
- `provider/authelia/provider.go`

**Dependencies**: Task 1.2

---

### Task 1.8: Update Main Entry Point

**Description**: Update main.go to use new configuration and router.

**Acceptance Criteria**:
- [ ] Update `main()` to call `LoadConfig()` with validation
- [ ] Handle configuration validation errors with descriptive messages
- [ ] Create `ProviderFactory` and get provider
- [ ] Create `Router` with base_path
- [ ] Log operation mode at startup
- [ ] Log base_path at startup
- [ ] Ensure graceful shutdown on configuration errors

**Files to Modify**:
- `main.go`

**Dependencies**: Task 1.3, Task 1.4, Task 1.6

---

### Task 1.9: Create Docker Compose Example

**Description**: Create deployment example for forward-auth mode with Traefik + Authelia.

**Acceptance Criteria**:
- [ ] Create `deployment/example/forward-auth/` directory
- [ ] Create `docker-compose.yml` with Traefik, Authelia, elastauth, Elasticsearch, Redis
- [ ] Create `config.yml` for elastauth with `operation_mode: forward-auth`
- [ ] Create `.env.example` with required environment variables
- [ ] Create Traefik configuration with forward auth middleware
- [ ] Create Authelia configuration with test users
- [ ] Create minimal README.md with quick start instructions
- [ ] Add link to comprehensive Starlight documentation
- [ ] Test end-to-end: Traefik → Authelia → elastauth → Elasticsearch

**Files to Create**:
- `deployment/example/forward-auth/docker-compose.yml`
- `deployment/example/forward-auth/config.yml`
- `deployment/example/forward-auth/.env.example`
- `deployment/example/forward-auth/traefik.yml`
- `deployment/example/forward-auth/authelia-config.yml`
- `deployment/example/forward-auth/authelia-users.yml`
- `deployment/example/forward-auth/README.md`

**Dependencies**: Task 1.8

---

### Task 1.10: Update Documentation

**Description**: Update documentation for forward-auth mode.

**Acceptance Criteria**:
- [ ] Create Starlight documentation page for forward-auth mode
- [ ] Include architecture diagram with Mermaid
- [ ] Include sequence diagram with Mermaid
- [ ] Document configuration options with Tabs (YAML vs env vars)
- [ ] Document deployment steps with Steps component
- [ ] Document when to use forward-auth mode with Cards
- [ ] Document security considerations with Asides
- [ ] Add troubleshooting section
- [ ] Add links to deployment example

**Files to Create**:
- `docs/src/content/docs/deployment/forward-auth-mode.mdx`

**Dependencies**: Task 1.9

---

### Task 1.11: Manual Testing

**Description**: Perform manual testing of forward-auth mode.

**Manual Testing Checklist**:
- [ ] Start docker-compose example
- [ ] Access Elasticsearch through Traefik (http://localhost/)
- [ ] Verify redirect to Authelia login page
- [ ] Login with test credentials (user: testuser, password: testpass)
- [ ] Verify redirect back to Elasticsearch
- [ ] Verify Elasticsearch response is returned
- [ ] Check elastauth logs for user creation
- [ ] Verify Elasticsearch user was created with correct roles
- [ ] Make second request, verify cached credentials are used
- [ ] Access health endpoint at http://localhost/elastauth/health
- [ ] Access config endpoint at http://localhost/elastauth/config
- [ ] Verify config endpoint shows `operation_mode: forward-auth`
- [ ] Stop and restart elastauth, verify configuration loads correctly
- [ ] Test with invalid configuration, verify descriptive error message

**Dependencies**: Task 1.9

---

### Task 1.12: Phase 1 Gate

**Description**: Verify Phase 1 completion criteria.

**Gate Checklist**:
- [ ] All Phase 1 tasks marked complete
- [ ] `go build` succeeds without errors or warnings
- [ ] `go test ./...` passes all tests
- [ ] Manual testing checklist 100% complete
- [ ] Docker Compose example works end-to-end
- [ ] Documentation updated and reviewed
- [ ] Configuration validation works correctly
- [ ] Existing Authelia functionality preserved
- [ ] Code committed with descriptive message

**Git Tag**: `operation-modes-phase-1-complete`

**Dependencies**: All Phase 1 tasks

---

## Phase 2: Direct Auth Mode (OIDC OAuth2)

### Task 2.1: Add OIDC Dependencies

**Description**: Add required Go libraries for OIDC authentication.

**Acceptance Criteria**:
- [ ] Add `github.com/coreos/go-oidc/v3/oidc` dependency
- [ ] Add `golang.org/x/oauth2` dependency
- [ ] Add `github.com/gorilla/sessions` dependency
- [ ] Run `go mod tidy`
- [ ] Verify dependencies download correctly

**Files to Modify**:
- `go.mod`
- `go.sum`

**Dependencies**: Phase 1 complete

---

### Task 2.2: Implement OIDC Configuration

**Description**: Complete OIDC configuration structure with all required fields.

**Acceptance Criteria**:
- [ ] Complete `OIDCConfig` struct with all fields from design
- [ ] Add `ClaimMappings` struct
- [ ] Add validator tags for required fields
- [ ] Add default values for optional fields
- [ ] Add environment variable support for all OIDC settings
- [ ] Update config.yml.example with OIDC section

**Files to Modify**:
- `libs/config.go`
- `config.yml.example`

**Dependencies**: Task 2.1

---

### Task 2.3: Implement OIDC Authentication Handler

**Description**: Implement OIDC authentication handler with OAuth2 authorization code flow.

**Acceptance Criteria**:
- [ ] Create `OIDCAuthHandler` struct
- [ ] Implement `NewOIDCAuthHandler()` with OIDC discovery
- [ ] Implement `CheckAuthentication()` for session validation
- [ ] Implement `InitiateLogin()` for OAuth2 redirect
- [ ] Implement `HandleCallback()` for code exchange
- [ ] Implement state parameter generation and validation
- [ ] Implement PKCE support (optional)
- [ ] Implement session cookie creation
- [ ] Add logging for authentication flow at debug level

**Files to Create**:
- `provider/oidc/handler.go`

**Dependencies**: Task 2.2

---

### Task 2.4: Implement Session Management

**Description**: Implement secure session management with cookies.

**Acceptance Criteria**:
- [ ] Use gorilla/sessions for session store
- [ ] Configure session cookies: HttpOnly, Secure, SameSite=Lax
- [ ] Store user info in session: username, email, groups, full_name
- [ ] Implement session expiration based on configured duration
- [ ] Encrypt session data using secret_key
- [ ] Implement session validation
- [ ] Add logging for session creation/validation at debug level

**Files to Modify**:
- `provider/oidc/handler.go`

**Dependencies**: Task 2.3

---

### Task 2.5: Implement OIDC Token Validation

**Description**: Implement OIDC ID token validation with JWKS.

**Acceptance Criteria**:
- [ ] Use coreos/go-oidc for token verification
- [ ] Fetch JWKS from OIDC provider
- [ ] Verify token signature using JWKS
- [ ] Verify token issuer matches configuration
- [ ] Verify token audience matches client_id
- [ ] Verify token expiration
- [ ] Extract claims using configured claim mappings
- [ ] Handle groups claim (string or array)
- [ ] Add logging for token validation at debug level

**Files to Modify**:
- `provider/oidc/handler.go`

**Dependencies**: Task 2.4

---

### Task 2.6: Implement OIDC Provider

**Description**: Implement OIDC provider that integrates with authentication handler.

**Acceptance Criteria**:
- [ ] Create `Provider` struct with OIDC handler
- [ ] Implement `NewProvider()` constructor
- [ ] Implement `GetUser()` that extracts user from session context
- [ ] Implement `Type()` that returns "oidc"
- [ ] Add logging for user extraction at debug level

**Files to Create**:
- `provider/oidc/provider.go`

**Dependencies**: Task 2.5

---

### Task 2.7: Update Provider Factory for OIDC

**Description**: Update provider factory to support OIDC provider.

**Acceptance Criteria**:
- [ ] Update `ProviderFactory` to initialize OIDC handler
- [ ] Update `GetProvider()` to return OIDC provider for `direct-auth` mode
- [ ] Add `GetOIDCHandler()` method for callback routing
- [ ] Validate that OIDC configuration is complete
- [ ] Add error handling for OIDC initialization failures

**Files to Modify**:
- `provider/factory.go`

**Dependencies**: Task 2.6

---

### Task 2.8: Update Router for OIDC Callback

**Description**: Update router to handle OIDC callback endpoint.

**Acceptance Criteria**:
- [ ] Add OIDC handler field to Router struct
- [ ] Update `handleInternalPath()` to route `/callback` to OIDC handler
- [ ] Only enable callback endpoint in `direct-auth` mode
- [ ] Add authentication check before proxying in `direct-auth` mode
- [ ] Store user info in request context after authentication
- [ ] Redirect to OIDC provider if no valid session

**Files to Modify**:
- `libs/router.go`

**Dependencies**: Task 2.7

---

### Task 2.9: Update Main Entry Point for OIDC

**Description**: Update main.go to initialize OIDC components.

**Acceptance Criteria**:
- [ ] Pass OIDC handler to Router constructor
- [ ] Handle OIDC initialization errors gracefully
- [ ] Log OIDC issuer at startup in `direct-auth` mode
- [ ] Validate that `proxy.enabled=true` in `direct-auth` mode

**Files to Modify**:
- `main.go`

**Dependencies**: Task 2.8

---

### Task 2.10: Create Docker Compose Example for OIDC

**Description**: Create deployment example for direct-auth mode with OIDC provider.

**Acceptance Criteria**:
- [ ] Create `deployment/example/direct-auth/` directory
- [ ] Create `docker-compose.yml` with Keycloak, elastauth, Elasticsearch, Redis
- [ ] Create `config.yml` for elastauth with `operation_mode: direct-auth`
- [ ] Create `.env.example` with required environment variables
- [ ] Configure Keycloak realm with test client and users
- [ ] Create minimal README.md with quick start instructions
- [ ] Add link to comprehensive Starlight documentation
- [ ] Test end-to-end: Browser → elastauth → OIDC → elastauth → Elasticsearch

**Files to Create**:
- `deployment/example/direct-auth/docker-compose.yml`
- `deployment/example/direct-auth/config.yml`
- `deployment/example/direct-auth/.env.example`
- `deployment/example/direct-auth/keycloak-realm.json`
- `deployment/example/direct-auth/README.md`

**Dependencies**: Task 2.9

---

### Task 2.11: Update Documentation for OIDC

**Description**: Update documentation for direct-auth mode.

**Acceptance Criteria**:
- [ ] Create Starlight documentation page for direct-auth mode
- [ ] Include architecture diagram with Mermaid
- [ ] Include sequence diagram with Mermaid (OAuth2 flow)
- [ ] Document configuration options with Tabs (YAML vs env vars)
- [ ] Document deployment steps with Steps component
- [ ] Document when to use direct-auth mode with Cards
- [ ] Document browser requirement with Aside
- [ ] Document security considerations with Asides
- [ ] Add troubleshooting section
- [ ] Add links to deployment example

**Files to Create**:
- `docs/src/content/docs/deployment/direct-auth-mode.mdx`

**Dependencies**: Task 2.10

---

### Task 2.12: Create Mode Comparison Documentation

**Description**: Create documentation comparing forward-auth and direct-auth modes.

**Acceptance Criteria**:
- [ ] Create Starlight documentation page for mode comparison
- [ ] Use CardGrid to compare features
- [ ] Use table to compare security models
- [ ] Use table to compare access method compatibility
- [ ] Document decision criteria for choosing a mode
- [ ] Add LinkCards to mode-specific documentation
- [ ] Include deployment architecture diagrams for both modes

**Files to Create**:
- `docs/src/content/docs/deployment/modes.mdx`

**Dependencies**: Task 2.11

---

### Task 2.13: Update Migration Guide

**Description**: Create migration guide for existing users.

**Acceptance Criteria**:
- [ ] Create migration guide documentation
- [ ] Document breaking change (operation_mode required)
- [ ] Provide migration steps for Authelia users
- [ ] Provide migration steps for OIDC users
- [ ] Document configuration changes
- [ ] Provide before/after configuration examples
- [ ] Document how to test after migration

**Files to Create**:
- `docs/src/content/docs/guides/migration-operation-modes.mdx`

**Dependencies**: Task 2.12

---

### Task 2.14: Manual Testing for OIDC

**Description**: Perform manual testing of direct-auth mode.

**Manual Testing Checklist**:
- [ ] Start docker-compose example with Keycloak
- [ ] Access Elasticsearch through elastauth (http://localhost:8080/)
- [ ] Verify redirect to Keycloak login page
- [ ] Login with test credentials
- [ ] Verify redirect to elastauth callback
- [ ] Verify redirect to original Elasticsearch URL
- [ ] Verify Elasticsearch response is returned
- [ ] Check elastauth logs for user creation
- [ ] Verify Elasticsearch user was created with correct roles
- [ ] Make second request, verify session is used (no redirect)
- [ ] Wait for session expiration, verify re-authentication
- [ ] Access health endpoint at http://localhost:8080/elastauth/health
- [ ] Access config endpoint at http://localhost:8080/elastauth/config
- [ ] Verify config endpoint shows `operation_mode: direct-auth`
- [ ] Test with invalid OIDC token, verify 401 response
- [ ] Test with invalid state parameter, verify 400 response
- [ ] Stop and restart elastauth, verify configuration loads correctly

**Dependencies**: Task 2.10

---

### Task 2.15: Regression Testing

**Description**: Verify Phase 1 (forward-auth) still works after Phase 2 implementation.

**Regression Testing Checklist**:
- [ ] Start Phase 1 docker-compose example
- [ ] Run all Phase 1 manual tests
- [ ] Verify forward-auth mode works unchanged
- [ ] Verify configuration validation still works
- [ ] Verify health/config endpoints work
- [ ] Verify Authelia provider works correctly

**Dependencies**: Task 2.14

---

### Task 2.16: Phase 2 Gate

**Description**: Verify Phase 2 completion criteria.

**Gate Checklist**:
- [ ] All Phase 2 tasks marked complete
- [ ] `go build` succeeds without errors or warnings
- [ ] `go test ./...` passes all tests
- [ ] Manual testing checklist 100% complete
- [ ] Regression testing checklist 100% complete
- [ ] Docker Compose example works end-to-end
- [ ] Documentation updated and reviewed
- [ ] OIDC OAuth2 flow works correctly
- [ ] Session management works correctly
- [ ] Phase 1 (forward-auth) still works
- [ ] Code committed with descriptive message

**Git Tag**: `operation-modes-phase-2-complete`

**Dependencies**: All Phase 2 tasks

---

## Post-Implementation Tasks

### Task 3.1: Update README

**Description**: Update main README with operation modes information.

**Acceptance Criteria**:
- [ ] Add operation modes section to README
- [ ] Add links to mode-specific documentation
- [ ] Update quick start examples for both modes
- [ ] Add migration notice for existing users
- [ ] Update feature list

**Files to Modify**:
- `README.md`

**Dependencies**: Phase 2 complete

---

### Task 3.2: Update CHANGELOG

**Description**: Document breaking changes and new features in CHANGELOG.

**Acceptance Criteria**:
- [ ] Add new version section to CHANGELOG
- [ ] Document breaking change: operation_mode required
- [ ] Document new feature: forward-auth mode
- [ ] Document new feature: direct-auth mode with OIDC
- [ ] Document new feature: configurable base_path
- [ ] Document migration guide link

**Files to Modify**:
- `CHANGELOG.md`

**Dependencies**: Phase 2 complete

---

### Task 3.3: Create Release

**Description**: Prepare release with version bump.

**Acceptance Criteria**:
- [ ] Bump version to next major version (breaking change)
- [ ] Create release notes
- [ ] Tag release in git
- [ ] Build release binaries
- [ ] Update GitHub release page

**Dependencies**: Task 3.1, Task 3.2

---

## Task Summary

### Phase 1: Forward Auth Mode
- **Total Tasks**: 12
- **Estimated Duration**: 2-3 days
- **Key Deliverables**: Configuration refactoring, forward-auth mode, docker-compose example

### Phase 2: Direct Auth Mode
- **Total Tasks**: 16
- **Estimated Duration**: 3-4 days
- **Key Deliverables**: OIDC OAuth2 flow, session management, docker-compose example

### Post-Implementation
- **Total Tasks**: 3
- **Estimated Duration**: 1 day
- **Key Deliverables**: Documentation updates, release preparation

### Total Project
- **Total Tasks**: 31
- **Estimated Duration**: 6-8 days
- **Key Deliverables**: Secure operation modes, comprehensive documentation, deployment examples
