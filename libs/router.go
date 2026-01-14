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
	// Check for special paths that bypass proxying
	if r.isSpecialPath(req.URL.Path) {
		// Use Echo server to handle special paths
		r.echoServer.ServeHTTP(w, req)
		return
	}

	// Route based on mode
	switch r.mode {
	case AuthOnlyMode:
		// In auth-only mode, use Echo server for all requests
		r.echoServer.ServeHTTP(w, req)
	case TransparentProxyMode:
		// In proxy mode, use goproxy for non-special paths
		r.proxyServer.ServeHTTP(w, req)
	default:
		http.Error(w, "Invalid operating mode", http.StatusInternalServerError)
	}
}

// isSpecialPath checks if the path should bypass proxying
// Special paths are handled directly by the Echo server regardless of mode
func (r *Router) isSpecialPath(path string) bool {
	specialPaths := []string{
		"/health",
		"/ready",
		"/live",
		"/config",
		"/docs",
		"/api/openapi.yaml",
		"/metrics",
	}

	for _, sp := range specialPaths {
		if strings.HasPrefix(path, sp) {
			return true
		}
	}
	return false
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
