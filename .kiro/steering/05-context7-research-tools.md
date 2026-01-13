---
inclusion: always
---

# Context7 MCP Tools for Research and Development

## Mandatory Context7 Usage

During research and development phases of elastauth development, you MUST use Context7 MCP tools to gather up-to-date information about libraries, frameworks, and technologies.

## When to Use Context7 Tools

### Research Phase Requirements
- **ALWAYS** use Context7 when researching new libraries or frameworks
- **ALWAYS** use Context7 when investigating integration approaches
- **ALWAYS** use Context7 when looking up current best practices
- **ALWAYS** use Context7 when checking for updated APIs or configuration patterns

### Specific Use Cases for elastauth

1. **Authentication Provider Research**
   - Research OAuth2/OIDC libraries and standards
   - Investigate provider-specific integration patterns
   - Look up current authentication best practices

2. **Framework and Library Research**
   - Research documentation frameworks (like Starlight)
   - Investigate caching libraries and patterns
   - Look up Go library best practices and updates

3. **Configuration and Deployment Research**
   - Research Kubernetes deployment patterns
   - Investigate GitHub Actions workflow best practices
   - Look up container and cloud deployment strategies

## Context7 Tool Usage Pattern

### Step 1: Resolve Library ID
```
Use mcp_context7_resolve_library_id to find the correct library
- Provide clear library name (e.g., "starlight", "oauth2", "oidc")
- Include context about your use case
- Select the most relevant result based on reputation and snippets
```

### Step 2: Query Documentation
```
Use mcp_context7_query_docs with the resolved library ID
- Ask specific questions about implementation
- Request configuration examples
- Look for best practices and common patterns
```

### Research Quality Standards

**Comprehensive Research**:
- Don't rely on outdated knowledge - always check current documentation
- Use Context7 to verify API changes and new features
- Research multiple approaches before making implementation decisions

**Documentation Integration**:
- Include Context7 findings in design decisions
- Reference current documentation in implementation comments
- Update elastauth documentation based on latest best practices

## elastauth-Specific Research Areas

### Priority Research Topics

1. **OAuth2/OIDC Standards**
   - Current OAuth2 and OIDC specifications
   - Go library recommendations and patterns
   - Security best practices and common vulnerabilities

2. **Documentation Frameworks**
   - Starlight features and configuration options
   - Static site generation best practices
   - GitHub Pages deployment patterns

3. **Caching and Performance**
   - Go caching library patterns
   - Redis integration best practices
   - Performance optimization techniques

4. **Kubernetes and Containerization**
   - Container best practices for Go applications
   - Kubernetes deployment patterns
   - Health check and monitoring approaches

### Research Documentation

**Document Findings**:
- Include Context7 research results in design documents
- Reference specific library versions and features
- Note any breaking changes or migration requirements

**Share Knowledge**:
- Update steering rules based on research findings
- Include research-based recommendations in code comments
- Document decision rationale based on current best practices

## Integration with Development Workflow

### Before Implementation
1. **Research Phase**: Use Context7 to understand current best practices
2. **Design Phase**: Incorporate research findings into design decisions
3. **Implementation Phase**: Reference current documentation during coding

### During Implementation
- Use Context7 to resolve specific implementation questions
- Look up current API patterns and examples
- Verify configuration options and parameters

### After Implementation
- Use Context7 to research testing best practices
- Look up deployment and monitoring patterns
- Research performance optimization techniques

## Quality Assurance

### Research Validation
- Cross-reference multiple sources when possible
- Verify information with official documentation
- Check for recent updates or breaking changes

### Implementation Validation
- Test implementations against current library versions
- Validate configuration against current documentation
- Ensure compatibility with latest best practices

## Remember

**Stay Current**: Technology moves fast. Context7 helps ensure elastauth uses current best practices and avoids deprecated patterns.

**Research First**: Before implementing any new feature or integration, research the current state of the technology using Context7.

**Document Decisions**: Include research findings in design documents and code comments to help future maintainers understand the rationale.

The goal is to build elastauth using the most current and appropriate technologies, patterns, and best practices available.