package cache

import (
	"strings"
)

// Public cache key prefixes used across handlers.
// Keep these in sync when introducing new public cache keys.
const (
	PrefixBootstrap   = "bootstrap:"
	PrefixArticles    = "articles:"  // list: articles:list:...
	PrefixArticle     = "article:"   // detail: article:{slug}
	PrefixPagesList   = "pages:list:"
	PrefixPage        = "page:" // page:{slug}:{locale}
	PrefixContent     = "content:"
	PrefixAnalytics   = "admin:analytics:"
	PrefixDashboard   = "admin:dashboard:"
)

// InvalidateBootstrap drops SPA bootstrap payloads.
func InvalidateBootstrap(c *Cache) {
	if c == nil {
		return
	}
	c.DeletePrefix(PrefixBootstrap)
}

// InvalidateArticlePublic drops public article list + optional detail by slug.
// Note: article detail keys are "article:{slug}" while lists use "articles:…".
// We never DeletePrefix("article:") alone, because that would also match "articles:".
func InvalidateArticlePublic(c *Cache, slug string) {
	if c == nil {
		return
	}
	c.DeletePrefix(PrefixArticles)
	if slug != "" {
		c.Delete(PrefixArticle + slug)
	} else {
		// Unknown slug: drop all article detail keys without touching "articles:" lists
		// (lists already cleared above).
		c.DeleteWhere(func(key string) bool {
			return strings.HasPrefix(key, PrefixArticle) && !strings.HasPrefix(key, PrefixArticles)
		})
	}
	// Bootstrap embeds theme/pages; article index may appear on home via client fetch only,
	// but keep bootstrap warm unless slug-less bulk ops.
	InvalidateBootstrap(c)
}

// InvalidatePagePublic drops public page list/detail/content caches for a slug.
func InvalidatePagePublic(c *Cache, slug string) {
	if c == nil {
		return
	}
	c.DeletePrefix(PrefixPagesList)
	InvalidateBootstrap(c)
	if slug != "" {
		c.DeletePrefix(PrefixPage + slug + ":")
		c.DeletePrefix(PrefixContent + slug + ":")
		return
	}
	c.DeletePrefix(PrefixPage)
	c.DeletePrefix(PrefixContent)
}

// InvalidateThemeOrSiteConfig drops bootstrap and global content config keys.
func InvalidateThemeOrSiteConfig(c *Cache) {
	if c == nil {
		return
	}
	InvalidateBootstrap(c)
	c.DeletePrefix(PrefixContent + "global:")
}

// InvalidatePublicFromContentEvent applies fine-grained invalidation based on event payload.
// contentType should be "article" or "page" (see eventbus.ContentEventPayload).
// Unknown types fall back to invalidating all known public prefixes (not a full Flush of
// admin keys that may share the same store).
func InvalidatePublicFromContentEvent(c *Cache, contentType, slug string) {
	if c == nil {
		return
	}
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "article":
		InvalidateArticlePublic(c, slug)
	case "page":
		InvalidatePagePublic(c, slug)
	default:
		InvalidateAllPublicPrefixes(c)
	}
}

// InvalidateAllPublicPrefixes clears all known public read-cache prefixes without
// wiping unrelated admin keys (analytics/dashboard) that may live in the same Cache.
func InvalidateAllPublicPrefixes(c *Cache) {
	if c == nil {
		return
	}
	c.DeletePrefix(PrefixBootstrap)
	c.DeletePrefix(PrefixArticles)
	c.DeleteWhere(func(key string) bool {
		return strings.HasPrefix(key, PrefixArticle) && !strings.HasPrefix(key, PrefixArticles)
	})
	c.DeletePrefix(PrefixPagesList)
	c.DeletePrefix(PrefixPage)
	c.DeletePrefix(PrefixContent)
}
