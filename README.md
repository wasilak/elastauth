# elastauth

[![Docker Repository on Quay](https://quay.io/repository/wasilak/elastauth/status "Docker Repository on Quay")](https://quay.io/repository/wasilak/elastauth) [![CI](https://github.com/wasilak/elastauth/actions/workflows/main.yml/badge.svg)](https://github.com/wasilak/elastauth/actions/workflows/main.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/wasilak/elastauth.svg)](https://pkg.go.dev/github.com/wasilak/elastauth)

A stateless authentication proxy for Elasticsearch and Kibana with pluggable authentication providers.

## What is elastauth?

elastauth bridges authentication systems with Elasticsearch/Kibana by managing temporary user credentials and providing seamless access without paid subscriptions. It supports two operating modes:

- **Authentication-Only Mode**: Works with reverse proxies like Traefik for forward authentication
- **Transparent Proxy Mode**: Handles both authentication and proxying in a single service

## Quick Start

```bash
# Pull and run with Docker
docker pull quay.io/wasilak/elastauth:latest
docker run -v ./config.yml:/config.yml -p 3000:3000 quay.io/wasilak/elastauth:latest

# Or clone and build
git clone --depth 1 https://github.com/wasilak/elastauth.git
cd elastauth
go build
./elastauth
```

See [config.yml.example](config.yml.example) for configuration options.

## Documentation

📚 **[Full Documentation](https://wasilak.github.io/elastauth/)** - Complete guides, configuration reference, and deployment examples

Quick links:
- [Getting Started](https://wasilak.github.io/elastauth/getting-started/concepts/)
- [Configuration Guide](https://wasilak.github.io/elastauth/configuration/)
- [Authentication Providers](https://wasilak.github.io/elastauth/providers/)
- [Deployment Examples](https://wasilak.github.io/elastauth/deployment/)

## Features

- **Dual Operating Modes** - Authentication-only or transparent proxy
- **Pluggable Providers** - Authelia, OIDC, Casdoor support
- **Stateless Operation** - No persistent authentication state
- **Credential Caching** - Redis, memory, or file-based caching
- **Horizontal Scaling** - Multi-instance deployments with Redis
- **Security First** - AES-256 encryption, input validation, secure defaults

## Built With

- [Go](https://golang.org/) - Core language
- [Echo](https://echo.labstack.com/) - HTTP framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [goproxy](https://github.com/elazarl/goproxy) - HTTP proxy functionality
- [Elasticsearch Go Client](https://github.com/elastic/go-elasticsearch) - Elasticsearch integration
- [Starlight](https://starlight.astro.build/) - Documentation site

## Contributing

Contributions welcome! Please check the [documentation](https://wasilak.github.io/elastauth/) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.
