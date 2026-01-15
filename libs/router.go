package libs

import (
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/labstack/echo/v4"
)

// OperatingMode represents the system's operating mode
type OperatingMode int

const (
	// AuthOnlyMode is the authentication-only mode (current behavior)
	AuthOnlyMode OperatingMode = iota
	// TransparentProxyMode is the transparent proxy mode (new)
	TransparentProxyMode
)

// Router handles request routing based on mode and path
type Router struct {
	mode        OperatingMode
	authHandler http.Handler
	proxyServer *goproxy.ProxyHttpServer
	echoServer  *echo.Echo
}

// NewRouter creates a new Router instance
func NewRouter(mode OperatingMode, echoServer *echo.Echo, proxyServer *goproxy.ProxyHttpServer) *Router {
	return &Router{
		mode:        mode,
		echoServer:  echoServer,
		proxyServer: proxyServer,
	}
}

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// All elastauth endpoints are scoped under /elastauth/*
	// Everything else (/*) goes to proxy in TransparentProxyMode
	if strings.HasPrefix(req.URL.Path, "/elastauth/") || req.URL.Path == "/elastauth" {
		// Use Echo server to handle elastauth endpoints
		r.echoServer.ServeHTTP(w, req)
		return
	}

	// Route based on mode
	switch r.mode {
	case AuthOnlyMode:
		// In auth-only mode, use Echo server for all requests
		r.echoServer.ServeHTTP(w, req)
	case TransparentProxyMode:
		// In proxy mode, proxy all non-elastauth paths to Elasticsearch
		r.proxyServer.ServeHTTP(w, req)
	default:
		http.Error(w, "Invalid operating mode", http.StatusInternalServerError)
	}
}

// GetMode returns the current operating mode
func (r *Router) GetMode() OperatingMode {
	return r.mode
}

// String returns a string representation of the operating mode
func (m OperatingMode) String() string {
	switch m {
	case AuthOnlyMode:
		return "auth-only"
	case TransparentProxyMode:
		return "proxy"
	default:
		return "unknown"
	}
}
