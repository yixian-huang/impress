package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

const (
	// SiteContextKey is the gin context key used to store the resolved site
	SiteContextKey ContextKey = "site"
)

// SiteResolver returns a middleware that resolves the current site from the
// request Host header (subdomain mode) or URL path prefix (subpath mode) and
// stores the *model.Site in the gin context under SiteContextKey.
//
// Resolution order:
//  1. Try to find a site whose SubPath matches the leading path component (subpath mode).
//  2. Fall back to matching the request Host (without port) against site.Domain.
//
// If no site is matched the request continues without site context — handlers
// that require a site should check GetSiteContext and respond accordingly.
func SiteResolver(siteRepo repository.SiteRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// --- Subpath mode: try to match a leading path prefix ---
		urlPath := c.Request.URL.Path
		if urlPath != "" && urlPath != "/" {
			// Extract the first path segment (e.g. "/blog/..." → "blog")
			trimmed := strings.TrimPrefix(urlPath, "/")
			segment := "/" + strings.SplitN(trimmed, "/", 2)[0]

			site, err := siteRepo.FindBySubPath(ctx, segment)
			if err == nil && site != nil {
				c.Set(string(SiteContextKey), site)
				c.Next()
				return
			}
		}

		// --- Subdomain / domain mode: match Host header ---
		host := c.Request.Host
		// Strip port if present
		if idx := strings.LastIndex(host, ":"); idx != -1 {
			host = host[:idx]
		}

		if host != "" {
			site, err := siteRepo.FindByDomain(ctx, host)
			if err == nil && site != nil {
				c.Set(string(SiteContextKey), site)
			}
		}

		c.Next()
	}
}

// GetSiteContext retrieves the resolved site from the gin context.
// Returns nil if no site was resolved for the current request.
func GetSiteContext(c *gin.Context) *model.Site {
	val, exists := c.Get(string(SiteContextKey))
	if !exists {
		return nil
	}
	site, ok := val.(*model.Site)
	if !ok {
		return nil
	}
	return site
}

// RequireSiteContext returns a middleware that aborts with 404 when no site
// has been resolved. Use this to protect routes that must be site-scoped.
func RequireSiteContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetSiteContext(c) == nil {
			c.JSON(404, gin.H{"error": gin.H{"message": "site not found"}})
			c.Abort()
			return
		}
		c.Next()
	}
}
