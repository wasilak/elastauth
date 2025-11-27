# Security Hardening Implementation Tasks

## Overview
These tasks implement unified security hardening across 6 interconnected components, addressing all 7 vulnerabilities identified in comprehensive security reviews.

---

## Phase 1: Input Validation Utilities

- [x] 1. Create validation utility functions in libs/utils.go
  - **File**: libs/utils.go
  - **Purpose**: Add validation functions for username, email, name, groups, and cache keys
  - **Scope**: Implement ValidateUsername, ValidateEmail, ValidateName, ParseAndValidateGroups, ValidateGroupName, EncodeForCacheKey
  - **Leverage**: `regexp` for patterns, `strings` for manipulation, existing utils.go structure
  - **Requirements**: 2 (Input Validation & Header Sanitization)
  - **Definition of Done**: 
    - All 6 validation functions implemented with clear error messages
    - Functions handle edge cases (empty strings, max length, special characters, whitespace)
    - All patterns match specification: username `^[a-zA-Z0-9._\-@]+$` with max 255 chars, email RFC 5322, name max 500 chars, groups with whitespace trimming
    - Unit tests written for each function covering valid/invalid cases
    - No external dependencies beyond stdlib

---

## Phase 2: Cryptographic Error Handling

- [x] 2. Refactor Encrypt function in libs/crypto.go to return errors
  - **File**: libs/crypto.go
  - **Purpose**: Replace panic-based error handling with proper error returns in Encrypt()
  - **Changes**: 
    - Change signature from `func Encrypt(ctx, stringToEncrypt, keyString string) (encryptedString string)` to `func Encrypt(ctx, stringToEncrypt, keyString string) (string, error)`
    - Handle `hex.DecodeString()` error for keyString
    - Handle `aes.NewCipher()` error
    - Handle `cipher.NewGCM()` error
    - Handle `io.ReadFull()` error for nonce generation
    - Return descriptive errors with context using `fmt.Errorf("...: %w", err)`
  - **Leverage**: Existing crypto algorithm (AES-256-GCM), nonce generation, OpenTelemetry tracing
  - **Requirements**: 1 (Cryptographic Error Handling), 5 (Error Handling Throughout)
  - **Definition of Done**:
    - Function returns (string, error) tuple
    - No panic() calls remain
    - All errors descriptive and wrapped with context
    - Existing algorithm and nonce behavior unchanged
    - Compiles without errors

- [x] 3. Refactor Decrypt function in libs/crypto.go to return errors
  - **File**: libs/crypto.go
  - **Purpose**: Replace panic-based error handling with proper error returns in Decrypt()
  - **Changes**:
    - Change signature from `func Decrypt(ctx, encryptedString, keyString string) (decryptedString string)` to `func Decrypt(ctx, encryptedString, keyString string) (string, error)`
    - Handle `hex.DecodeString()` error for keyString
    - Handle `hex.DecodeString()` error for encryptedString
    - Add bounds check: `if len(enc) < nonceSize + 16 { return "", fmt.Errorf(...) }`
    - Handle `aes.NewCipher()` error
    - Handle `cipher.NewGCM()` error
    - Handle `aesGCM.Open()` error with clear message about tampering/invalid ciphertext
    - Return descriptive errors with context
  - **Leverage**: Existing crypto algorithm, OpenTelemetry tracing
  - **Requirements**: 1 (Cryptographic Error Handling), 5 (Error Handling Throughout)
  - **Definition of Done**:
    - Function returns (string, error) tuple
    - No panic() calls remain
    - Bounds checking prevents slice panics
    - All error paths tested
    - Compiles without errors

- [x] 4. Update all Encrypt/Decrypt call sites in libs/routes.go
  - **File**: libs/routes.go
  - **Purpose**: Handle new error returns from Encrypt() and Decrypt() functions
  - **Call Sites**: MainRoute line 121 (Encrypt), MainRoute line 178 (Decrypt)
  - **Changes**:
    - Line 121: Change `encryptedPassword := Encrypt(...)` to `encryptedPassword, err := Encrypt(...)`
    - Add error check after Encrypt: return sanitized error to client with HTTP 500
    - Line 178: Change `decryptedPassword := Decrypt(...)` to `decryptedPassword, err := Decrypt(...)`
    - Add error check after Decrypt: return sanitized error to client with HTTP 500
    - Log sanitized errors internally, return generic message to client
  - **Leverage**: Existing error handling patterns, slog for logging
  - **Requirements**: 1, 5
  - **Definition of Done**:
    - Both call sites properly handle error returns
    - No panics on crypto failures
    - Client receives HTTP 500 with generic message
    - Internal logging has full error context
    - All paths tested

---

## Phase 3: Configuration Validation

- [x] 5. Create configuration validation functions in libs/config.go
  - **File**: libs/config.go
  - **Purpose**: Add ValidateSecretKey() and ValidateRequiredConfig() functions
  - **Functions**:
    - ValidateSecretKey(key string) error: Check non-empty, valid hex, exactly 64 chars (32 bytes)
    - ValidateRequiredConfig(ctx) error: Check elasticsearch_host, elasticsearch_username, elasticsearch_password, secret_key all present
    - ValidateConfiguration(ctx) error: Comprehensive validation including cache_type, log_level, optional redis_host
  - **Leverage**: Existing Viper configuration, error wrapping patterns
  - **Requirements**: 4 (Configuration Validation at Startup)
  - **Definition of Done**:
    - All validation functions return clear error messages with env var names
    - Secret key validation checks hex format and 64-char length
    - Required config check lists missing fields with env var names
    - Comprehensive validation handles optional vs required fields correctly
    - Unit tests for each validation scenario

- [x] 6. Update HandleSecretKey() in libs/config.go to not log key
  - **File**: libs/config.go  
  - **Purpose**: Remove secret key from logs, prevent logging during key generation
  - **Changes**:
    - Line 88: Change `slog.String("key", key)` to not log the actual key value
    - Option 1: Log only key hash: `slog.String("keyHash", fmt.Sprintf("%x", sha256.Sum256([]byte(key))[:8]))`
    - Option 2: Log only that key was generated without the value
    - Update log level from Info to Warn to indicate action needed
    - Remove `fmt.Println(key)` from line 78, keep stdout to stderr and add instructions
  - **Leverage**: Existing logging patterns, crypto/sha256
  - **Requirements**: 3 (Logging Sanitization & Secret Protection)
  - **Definition of Done**:
    - Actual key never appears in logs
    - Log indicates key was generated with clear message
    - Unit tests verify no key in logs at any level

---

## Phase 4: Logging Sanitization Utilities

- [x] 7. Create logging sanitization functions in libs/utils.go
  - **File**: libs/utils.go
  - **Purpose**: Add SanitizeForLogging() and SafeLogError() utilities
  - **Functions**:
    - SanitizeForLogging(data interface{}) interface{}: Recursively redact sensitive fields in maps/structs
    - SafeLogError(err error) string: Create generic error message without exposing internals
    - IsSensitiveField(fieldName string) bool: Helper to identify sensitive field names
  - **Sensitive Field Names**: "password", "secret", "key", "token", "credential"
  - **Leverage**: `reflect` package for generic handling, `strings` for field matching
  - **Requirements**: 3 (Logging Sanitization & Secret Protection)
  - **Definition of Done**:
    - Sanitization handles maps, structs, slices recursively
    - Sensitive fields redacted as `***REDACTED***`
    - Non-sensitive fields preserved unchanged
    - SafeLogError creates generic message
    - Unit tests verify sensitive fields properly redacted

---

## Phase 5: Header Validation in Routes

- [x] 8. Add input validation to MainRoute in libs/routes.go
  - **File**: libs/routes.go
  - **Purpose**: Validate user-provided headers at request entry point
  - **Changes**:
    - After line 63 (get username): Add `if err := ValidateUsername(user); err != nil { return c.JSON(http.StatusBadRequest, ...) }`
    - At line 79 (parse groups): Replace with `userGroups, err := ParseAndValidateGroups(c.Request().Header.Get(headerName), enableWhitelist, whitelist)`
    - Check returned error and return HTTP 400 if validation fails
    - Optional: Add email validation if enabled, name validation if enabled
  - **Leverage**: New validation functions from utils.go, existing error handling
  - **Requirements**: 2 (Input Validation & Header Sanitization)
  - **Definition of Done**:
    - All header inputs validated before use
    - Invalid input returns HTTP 400 with descriptive message
    - Validation errors don't crash application
    - Whitelist validation optional and configurable
    - Unit/integration tests verify validation works

---

## Phase 6: Error Handling in Elasticsearch

- [x] 9. Fix unhandled JSON decode errors in libs/elastic.go
  - **File**: libs/elastic.go
  - **Purpose**: Capture and handle JSON decode errors
  - **Changes**:
    - Line 112: Change `json.NewDecoder(resp.Body).Decode(&body)` to `err := json.NewDecoder(resp.Body).Decode(&body)` followed by `if err != nil { return err }`
    - Line 151: Same change for second decode
    - Wrap errors with context about what failed
  - **Leverage**: Existing error handling patterns
  - **Requirements**: 5 (Error Handling Throughout)
  - **Definition of Done**:
    - JSON decode errors captured and returned
    - Errors wrapped with context
    - No silent failures
    - Unit tests verify error handling

- [x] 10. Sanitize Elasticsearch response logging in libs/elastic.go
  - **File**: libs/elastic.go
  - **Purpose**: Redact sensitive fields from Elasticsearch response logs
  - **Changes**:
    - Line 114: Change `slog.DebugContext(ctx, "Request response", slog.Any("body", body))` to use sanitized version
    - Line 157: Same change
    - Use new SanitizeForLogging() utility before logging
  - **Leverage**: New sanitization utilities from utils.go
  - **Requirements**: 3 (Logging Sanitization & Secret Protection)
  - **Definition of Done**:
    - Elasticsearch response logs use sanitized data
    - Sensitive fields (password, secret, etc.) redacted
    - Non-sensitive fields visible for debugging
    - Unit tests verify sanitization in logs

---

## Phase 7: Cache Key Security

- [x] 11. Update cache key generation in libs/routes.go
  - **File**: libs/routes.go
  - **Purpose**: Properly encode username in cache key
  - **Changes**:
    - Line 87: Change `cacheKey := "elastauth-" + user` to `cacheKey := "elastauth-" + url.QueryEscape(user)`
    - Import `net/url` if not present
  - **Leverage**: stdlib `net/url.QueryEscape()`
  - **Requirements**: 6 (Cache Key Security)
  - **Definition of Done**:
    - Cache keys properly encoded with QueryEscape
    - Special characters in usernames handled correctly
    - No cache key collisions from special chars
    - Unit tests verify encoding

---

## Phase 8: Configuration Startup Validation

- [x] 12. Add configuration validation call to main.go
  - **File**: main.go
  - **Purpose**: Validate configuration at startup before server starts
  - **Changes**:
    - After `libs.InitConfiguration()` call: Add `if err := libs.ValidateConfiguration(ctx); err != nil { log and exit }`
    - After `libs.HandleSecretKey(ctx)` call: Optional - HandleSecretKey now validates internally
    - Before `libs.WebserverInit(ctx)` call: Ensure validation passes
    - Exit with code 1 on validation failure, log clear error message
  - **Leverage**: New validation functions, existing error patterns
  - **Requirements**: 4 (Configuration Validation at Startup)
  - **Definition of Done**:
    - Configuration validated before server starts
    - Application exits with code 1 on validation failure
    - Error message includes env var name to fix issue
    - No server starts with invalid configuration
    - Unit/integration tests verify

---

## Phase 9: Comprehensive Unit Tests

- [ ] 13. Write unit tests for validation functions (utils_test.go)
  - **File**: libs/utils_test.go or new test file
  - **Purpose**: Test all validation functions with edge cases
  - **Coverage**:
    - ValidateUsername: valid/invalid formats, length boundaries, special characters
    - ValidateEmail: valid/invalid formats (if enabled)
    - ValidateName: length boundaries, control characters
    - ParseAndValidateGroups: single/multiple groups, whitespace trimming, empty values
    - ValidateGroupName: special characters, valid group names
    - EncodeForCacheKey: special characters properly encoded, collisions prevented
  - **Leverage**: Go testing package, existing test structure
  - **Requirements**: All
  - **Definition of Done**:
    - All validation functions have >90% test coverage
    - Edge cases and boundaries tested
    - Both success and failure paths tested
    - Tests run independently without side effects

- [ ] 14. Write unit tests for crypto error handling (crypto_test.go)
  - **File**: libs/crypto_test.go or new test file
  - **Purpose**: Test Encrypt/Decrypt error handling without panics
  - **Coverage**:
    - Encrypt with valid key/plaintext succeeds
    - Encrypt with invalid hex key returns error
    - Encrypt with short key returns error
    - Decrypt with valid ciphertext succeeds
    - Decrypt with invalid hex returns error
    - Decrypt with tampered ciphertext returns error
    - Decrypt with too-short ciphertext returns bounds check error
    - No panic() calls in any error path
  - **Leverage**: Go testing, existing test patterns
  - **Requirements**: 1, 5
  - **Definition of Done**:
    - All error paths tested
    - No panics triggered by invalid input
    - Error messages descriptive
    - >90% coverage of crypto functions

- [ ] 15. Write unit tests for config validation (config_test.go)
  - **File**: libs/config_test.go or new test file
  - **Purpose**: Test configuration validation functions
  - **Coverage**:
    - ValidateSecretKey with valid 64-char hex passes
    - ValidateSecretKey with non-hex fails
    - ValidateSecretKey with wrong length fails
    - ValidateRequiredConfig with all fields set passes
    - ValidateRequiredConfig with missing field fails with env var name
    - ValidateConfiguration comprehensive check
  - **Leverage**: Go testing, Viper mocking
  - **Requirements**: 4
  - **Definition of Done**:
    - All validation paths tested
    - Error messages include env var names
    - >90% coverage

- [ ] 16. Write unit tests for sanitization utilities (utils_test.go)
  - **File**: libs/utils_test.go
  - **Purpose**: Test logging sanitization functions
  - **Coverage**:
    - SanitizeForLogging redacts "password" field
    - SanitizeForLogging redacts "secret_key", "token"
    - SanitizeForLogging preserves non-sensitive fields
    - SanitizeForLogging handles nested maps/structs
    - SafeLogError creates generic message
    - IsSensitiveField correctly identifies sensitive names
  - **Leverage**: Go testing, reflect package
  - **Requirements**: 3
  - **Definition of Done**:
    - Sensitive fields properly redacted
    - Non-sensitive data preserved
    - Nested structures handled
    - >90% coverage

---

## Phase 10: Integration Testing and Verification

- [ ] 17. Write integration tests for request validation flow
  - **File**: tests/integration/ or new file
  - **Purpose**: Test complete request flow with validation
  - **Scenarios**:
    - Valid request → validation passes → succeeds
    - Invalid username → validation fails → HTTP 400
    - Invalid group → validation fails → HTTP 400
    - Valid request → encrypt succeeds → caches → decrypt succeeds
    - Request with tampered cache → decrypt fails → HTTP 500
  - **Leverage**: Existing test utilities, test fixtures
  - **Requirements**: 2, 1, 5
  - **Definition of Done**:
    - All request flow scenarios tested
    - Validation prevents invalid data reaching backend
    - Error handling returns appropriate status codes
    - No crashes on malformed input

- [ ] 18. Write integration tests for configuration startup
  - **File**: tests/integration/ or new file
  - **Purpose**: Test complete startup flow with validation
  - **Scenarios**:
    - Start with valid config → startup succeeds
    - Start with missing elasticsearch_host → startup fails with env var name
    - Start with invalid secret_key → startup fails with format requirement
    - Start with redis cache but no redis_host → startup fails
    - Application exits with code 1 on config error
  - **Leverage**: Test utilities, subprocess testing
  - **Requirements**: 4
  - **Definition of Done**:
    - Configuration validation prevents startup with bad config
    - Error messages helpful for debugging
    - Application exits with correct code
    - No server starts with invalid configuration

- [ ] 19. Verify no secrets in logs at any level
  - **File**: tests/integration/ or new file
  - **Purpose**: Audit logs to verify secrets never logged
  - **Verification**:
    - Capture logs at all levels (debug, info, warn, error)
    - Search for secret_key, elasticsearch_password, auth tokens
    - Verify no credentials in error responses to clients
    - Verify no credentials in internal error logging
    - Run log audit in test environment
  - **Leverage**: Log capture utilities, grep/search
  - **Requirements**: 3
  - **Definition of Done**:
    - No secrets found in logs at any level
    - Sensitive fields redacted in all output
    - Log audit passes completely

- [ ] 20. Run lint and type checking
  - **File**: Multiple (all modified files)
  - **Purpose**: Ensure code quality and correctness
  - **Commands**:
    - `go fmt ./...` - Format code
    - `go vet ./...` - Vet for common issues
    - `golangci-lint run ./...` - Comprehensive linting (if available)
  - **Requirements**: All
  - **Definition of Done**:
    - All files pass formatting
    - No vet errors or warnings
    - No linting errors
    - Code compiles without errors

---

## Success Criteria

✅ All 7 vulnerabilities addressed from security reviews  
✅ Cryptographic functions return errors instead of panicking  
✅ No secrets logged at any log level  
✅ All required configuration validated at startup  
✅ Input validation prevents malformed headers from reaching Elasticsearch  
✅ Application handles errors gracefully with appropriate HTTP status codes  
✅ Unit tests cover validation, error handling, and sanitization (>90% coverage)  
✅ No panic() statements remain in production code paths  
✅ All code passes linting and type checking  
