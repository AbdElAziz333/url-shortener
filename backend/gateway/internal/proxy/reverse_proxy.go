package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"aziz.dev/gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

const userIDHeader = "User-ID"

// ReverseProxy forwards requests to target, rewriting the fromPrefix of the
// path to toPrefix and passing the authenticated user ID downstream.
func ReverseProxy(fromPrefix, toPrefix, target string) (gin.HandlerFunc, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid service url %q: %w", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalHost := req.Host
		director(req)

		if strings.HasPrefix(req.URL.Path, fromPrefix) {
			req.URL.Path = toPrefix + strings.TrimPrefix(req.URL.Path, fromPrefix)
		}
		req.Header.Set("X-Forwarded-Host", originalHost)
		req.Host = targetURL.Host
	}

	return func(c *gin.Context) {
		// Never trust a client-supplied User-ID; only the auth middleware may set it.
		c.Request.Header.Del(userIDHeader)
		if id := c.GetString(middleware.UserIDKey); id != "" {
			c.Request.Header.Set(userIDHeader, id)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}