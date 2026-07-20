package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetWithTTLAndGet(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()

	c.SetWithTTL("k", "v", 50*time.Millisecond)
	v, ok := c.Get("k")
	require.True(t, ok)
	require.Equal(t, "v", v)

	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("k")
	require.False(t, ok)
}

func TestDeleteWhere(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()
	c.Set("a", 1)
	c.Set("b", 2)
	c.DeleteWhere(func(key string) bool { return key == "a" })
	_, okA := c.Get("a")
	_, okB := c.Get("b")
	require.False(t, okA)
	require.True(t, okB)
}

func TestInvalidateArticleDoesNotClobberUnrelatedPages(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()
	c.Set("article:hello", "detail")
	c.Set("articles:list:1:10::", "list")
	c.Set("bootstrap:zh:", "boot")
	c.Set("page:home:zh", "page")
	c.Set(PrefixAnalytics+"summary", "analytics")

	InvalidateArticlePublic(c, "hello")

	_, okDetail := c.Get("article:hello")
	_, okList := c.Get("articles:list:1:10::")
	_, okBoot := c.Get("bootstrap:zh:")
	_, okPage := c.Get("page:home:zh")
	_, okAnalytics := c.Get(PrefixAnalytics + "summary")

	require.False(t, okDetail)
	require.False(t, okList)
	require.False(t, okBoot)
	// Unrelated page detail must survive article invalidation
	require.True(t, okPage)
	// Admin analytics key must not be wiped by public invalidation
	require.True(t, okAnalytics)
}

func TestInvalidatePagePublicScoped(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()
	c.Set("page:about:zh", 1)
	c.Set("page:home:zh", 2)
	c.Set("content:about:zh", 3)
	c.Set("pages:list:zh", 4)
	c.Set("articles:list:1", 5)

	InvalidatePagePublic(c, "about")

	_, okAbout := c.Get("page:about:zh")
	_, okHome := c.Get("page:home:zh")
	_, okContent := c.Get("content:about:zh")
	_, okList := c.Get("pages:list:zh")
	_, okArticles := c.Get("articles:list:1")

	require.False(t, okAbout)
	require.True(t, okHome)
	require.False(t, okContent)
	require.False(t, okList)
	require.True(t, okArticles)
}

func TestInvalidatePublicFromContentEvent(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()
	c.Set("article:x", 1)
	c.Set("articles:list:1", 2)
	c.Set("page:y:zh", 3)

	InvalidatePublicFromContentEvent(c, "article", "x")
	_, okArt := c.Get("article:x")
	_, okList := c.Get("articles:list:1")
	_, okPage := c.Get("page:y:zh")
	require.False(t, okArt)
	require.False(t, okList)
	require.True(t, okPage)

	c.Set("page:y:zh", 3)
	c.Set("pages:list:zh", 4)
	InvalidatePublicFromContentEvent(c, "page", "y")
	_, okPage2 := c.Get("page:y:zh")
	_, okPL := c.Get("pages:list:zh")
	require.False(t, okPage2)
	require.False(t, okPL)
}

func TestInvalidateArticleWithoutSlugClearsArticleDetailsOnly(t *testing.T) {
	c := New(time.Minute)
	defer c.Stop()
	c.Set("article:a", 1)
	c.Set("article:b", 2)
	c.Set("articles:list:1", 3)

	InvalidateArticlePublic(c, "")

	_, okA := c.Get("article:a")
	_, okB := c.Get("article:b")
	_, okList := c.Get("articles:list:1")
	require.False(t, okA)
	require.False(t, okB)
	require.False(t, okList)
}
