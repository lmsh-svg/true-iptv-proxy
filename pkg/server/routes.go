package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Config struct {
	// Add any necessary fields, e.g., proxy config
}

// Serve sets up all routes
func (c *Config) Serve(r *gin.Engine) {
	c.routes(r)
}

// routes registers all routes
func (c *Config) routes(r *gin.Engine) {
	// Base group
	group := r.Group("/")

	// M3U endpoints
	group.GET("iptv.m3u", c.getM3U)
	group.POST("iptv.m3u", c.getM3U)

	// Reverse proxy routes for HLS / .ts / .m3u8 / live.php
	// Only numeric index in path, everything else handled dynamically
	for i := 0; i < 20; i++ {
		// Matches /{space}/{user}/{session}/{i}/{file}
		group.GET(fmt.Sprintf("/:space/:user/:session/%d/:file", i), c.reverseProxy)
	}
}

// getM3U serves M3U playlist
func (c *Config) getM3U(ctx *gin.Context) {
	// Implement your logic to return the M3U playlist
	ctx.String(http.StatusOK, "#EXTM3U\n# Playlist content here")
}

// reverseProxy dynamically forwards request
func (c *Config) reverseProxy(ctx *gin.Context) {
	space := ctx.Param("space")
	user := ctx.Param("user")
	session := ctx.Param("session")
	index := ctx.Param("file") // the wildcard file segment

	// Preserve the original query string (MAC, stream, token, extension)
	query := ctx.Request.URL.RawQuery

	// Construct the upstream URL
	targetURL := fmt.Sprintf("http://upstream.server/%s/%s/%s/%s?%s", space, user, session, index, query)

	// Forward request to upstream (simplified)
	resp, err := http.Get(targetURL)
	if err != nil {
		ctx.String(http.StatusBadGateway, fmt.Sprintf("Proxy error: %v", err))
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		ctx.Header(k, v[0])
	}

	// Copy status code
	ctx.Status(resp.StatusCode)

	// Copy body
	ctx.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, resp.Body)
		return err == nil
	})
}
