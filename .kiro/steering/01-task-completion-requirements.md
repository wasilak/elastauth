# Task Completion Requirements

When implementing tasks from elastauth specs, you MUST follow these completion criteria:

## Critical Thinking After Implementation

- After completing any task implementation, ALWAYS critically evaluate whether the code is functional or dead code
- Ask yourself: "Is this code actually being used anywhere in the elastauth application?"
- Verify integration points: Check if the implemented functionality is imported and called
- Search the elastauth codebase for actual usage of new structs, functions, or packages
- If code is not integrated, identify ALL places where it should be used and integrate it fully
- Don't just implement infrastructure - ensure it's wired into the application flow
- Reference your steering rules to ensure you're following Go best practices for elastauth
- Ensure new auth providers are properly registered and can be selected via configuration

## Build Verification for elastauth

- Always run `go build` or equivalent build command after implementing code changes
- Fix ALL build errors before marking a task as complete
- Address ALL build warnings before marking a task as complete
- A task is NOT complete if the build fails or produces warnings
- Verify that elastauth binary builds successfully with new changes
- Test that configuration loading works with new provider options

## Test Verification

- Run `go test ./...` to verify unit tests pass
- Fix any failing tests before marking task complete
- Add unit tests for new functionality where appropriate
- Test new auth providers with mock requests
- Verify cache integration works with new providers
- Test Elasticsearch integration remains functional

## Git Commit Requirement

- After each task is successfully implemented and verified, AUTOMATICALLY commit the changes
- Use a descriptive commit message that includes:
  - Type prefix (feat:, fix:, refactor:, etc.)
  - Brief description of what was implemented
  - Reference to the task number or name
- Stage all relevant files before committing
- Example: `feat: implement pluggable auth provider interface\n\nTask: 1.2 Create AuthProvider interface and factory`
- Commit immediately after verification steps pass, do not wait for user approval

## Verification Steps for elastauth

1. Implement the code changes
2. Run `go build` to verify the build succeeds
3. Run `go test ./...` to verify tests pass
4. Test configuration loading with new options
5. Verify elastauth starts successfully with new configuration
6. Fix any errors or warnings that appear
7. Re-run build and tests to confirm all issues are resolved
8. Mark the task as complete
9. AUTOMATICALLY commit the changes with a descriptive message

## elastauth-Specific Integration Points

- Ensure new providers integrate with existing `libs/routes.go` MainRoute
- Verify cache integration works with `cache/cache.go` interface
- Test Elasticsearch integration with `libs/elastic.go`
- Ensure configuration loading works with `libs/config.go` and Viper
- Verify logging integration with existing slog usage
- Test that existing Authelia functionality remains unbroken

## Why This Matters

- Ensures code integrates properly with the existing elastauth codebase
- Catches Go compilation errors, import issues, and configuration problems early
- Maintains production-ready code quality for elastauth
- Prevents broken builds from being committed
- Creates a clear history of elastauth development progress
- Makes it easy to track progress and revert changes if needed
- Ensures backward compatibility with existing elastauth deployments