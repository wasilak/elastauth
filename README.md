# elastauth

[![Docker Repository on Quay](https://quay.io/repository/wasilak/elastauth/status "Docker Repository on Quay")](https://quay.io/repository/wasilak/elastauth) [![CI](https://github.com/wasilak/elastauth/actions/workflows/main.yml/badge.svg)](https://github.com/wasilak/elastauth/actions/workflows/main.yml) [![Maintainability](https://api.codeclimate.com/v1/badges/d75cc6b44c7c33f0b530/maintainability)](https://codeclimate.com/github/wasilak/elastauth/maintainability) [![Go Reference](https://pkg.go.dev/badge/github.com/wasilak/elastauth.svg)](https://pkg.go.dev/github.com/wasilak/elastauth)

**A stateless authentication proxy for Elasticsearch and Kibana with pluggable authentication providers.**

elastauth bridges authentication systems with Elasticsearch/Kibana by:
- Extracting user information from various authentication providers
- Managing temporary Elasticsearch user credentials
- Providing seamless access to Kibana without paid subscriptions

## 🚀 Quick Start

### Authelia (Header-based)
```yaml
auth_provider: "authelia"
elasticsearch:
  hosts: ["http://localhost:9200"]
  username: "elastauth"
  password: "your-password"
default_roles: ["kibana_user"]
secret_key: "your-32-character-secret-key"
```

### OAuth2/OIDC (JWT tokens)
```yaml
auth_provider: "oidc"
oidc:
  issuer: "https://your-provider.com"
  client_id: "elastauth"
  client_secret: "your-secret"
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
elasticsearch:
  hosts: ["http://localhost:9200"]
  username: "elastauth"
  password: "your-password"
default_roles: ["kibana_user"]
secret_key: "your-32-character-secret-key"
```

## 📖 Documentation

### Core Concepts
- **[Architecture Overview](docs/concepts.md)** - How elastauth works
- **[API Documentation](docs/openapi.yaml)** - OpenAPI specification
- **[Interactive API Docs](/docs)** - Swagger UI (when running)

### Authentication Providers
- **[Authelia Provider](docs/providers/authelia.md)** - Header-based authentication
- **[OAuth2/OIDC Provider](docs/providers/oidc.md)** - JWT token authentication
  - Supports: Keycloak, Casdoor, Authentik, Auth0, Azure AD, Pocket-ID, Ory Hydra

### Caching & Performance
- **[Cache Providers](docs/cache/README.md)** - Memory, Redis, File caching
- **[Redis Cache](docs/cache/redis.md)** - Distributed caching for scaling
- **[Horizontal Scaling](docs/concepts.md#deployment-patterns)** - Multi-instance deployments

### Configuration & Deployment
- **[Configuration Examples](docs/examples/README.md)** - Complete setup examples
- **[Kubernetes Deployment](docs/examples/kubernetes.md)** - Production-ready K8s setup
- **[Docker Compose](docs/examples/docker-compose.md)** - Local development stack
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

## 🔧 Supported Authentication Systems

### Header-based (Authelia Provider)
- **[Authelia](https://www.authelia.com/)** - Popular authentication server
- **Traefik Forward Auth** - Any system that sets user headers
- **Custom Headers** - Configurable header names

### OAuth2/OIDC (Generic Provider)
- **[Keycloak](https://www.keycloak.org/)** - Open source identity management
- **[Casdoor](https://casdoor.org/)** - Web-based identity management  
- **[Authentik](https://goauthentik.io/)** - Modern authentication platform
- **[Auth0](https://auth0.com/)** - Cloud identity platform
- **[Azure AD](https://azure.microsoft.com/services/active-directory/)** - Microsoft identity
- **[Pocket-ID](https://github.com/stonith404/pocket-id)** - Self-hosted identity provider
- **[Ory Hydra](https://www.ory.sh/hydra/)** - OAuth2 and OpenID Connect server

## 🏗️ Architecture

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   elastauth     │    │  Elasticsearch  │
│   (Browser/App) │    │   Proxy         │    │   Cluster       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │ 1. Auth Request       │                       │
         │ (Headers/JWT/etc)     │                       │
         ├──────────────────────►│                       │
         │                       │ 2. Create/Update User │
         │                       ├──────────────────────►│
         │                       │                       │
         │ 3. Authorization      │ 4. User Created       │
         │    Header             │◄──────────────────────┤
         │◄──────────────────────┤                       │
         │                       │                       │
         │ 5. Access Elasticsearch/Kibana               │
         ├───────────────────────────────────────────────►│
```

**Key Features:**
- **Stateless Operation** - No persistent authentication state
- **Pluggable Providers** - Support for multiple authentication systems
- **Credential Caching** - Encrypted temporary password caching
- **Role Mapping** - Map user groups to Elasticsearch roles
- **Horizontal Scaling** - Multi-instance support with Redis cache

## 🐳 Docker

```bash
# Pull the image
docker pull quay.io/wasilak/elastauth:latest

# Run with configuration
docker run -v ./config.yml:/config.yml -p 5000:5000 quay.io/wasilak/elastauth:latest
```

## 🔒 Security

- **Credential Encryption** - All cached credentials encrypted with AES
- **Input Validation** - All user input validated and sanitized  
- **Secure Defaults** - Security-first configuration defaults
- **No Password Storage** - Temporary passwords only, automatically rotated

## 📊 Monitoring

- **Health Checks** - `/health` endpoint for load balancers
- **Configuration Info** - `/config` endpoint (sensitive values masked)
- **Structured Logging** - JSON logs with request correlation
- **OpenTelemetry** - Distributed tracing support

## 🤝 Contributing

1. **Issues** - Report bugs or request features via GitHub Issues
2. **Pull Requests** - Contributions welcome following Go best practices
3. **Documentation** - Help improve documentation and examples

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Need Help?** Check the [troubleshooting guide](docs/troubleshooting.md) or [open an issue](https://github.com/wasilak/elastauth/issues).
