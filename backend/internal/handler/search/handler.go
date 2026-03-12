package search

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/provider"
)

type Handler struct {
	search provider.SearchProvider
}

func NewHandler(search provider.SearchProvider) *Handler {
	return &Handler{search: search}
}

// PublicSearch performs a full-text search.
// @Summary      Search content
// @Description  Full-text search across articles and pages
// @Tags         Search
// @Produce      json
// @Param        q        query string true  "Search query"
// @Param        locale   query string false "Locale (zh or en)" default(zh)
// @Param        type     query string false "Content type filter"
// @Param        page     query int    false "Page number"    default(1)
// @Param        pageSize query int    false "Items per page" default(10)
// @Success      200 {object} object
// @Router       /public/search [get]
func (h *Handler) PublicSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})
		return
	}
	locale := c.DefaultQuery("locale", "zh")
	contentType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	resp, err := h.search.Search(c.Request.Context(), query, locale, contentType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PublicSuggest returns search suggestions.
// @Summary      Search suggestions
// @Description  Returns autocomplete suggestions for a search prefix
// @Tags         Search
// @Produce      json
// @Param        q      query string true  "Search prefix"
// @Param        locale query string false "Locale (zh or en)" default(zh)
// @Param        limit  query int    false "Max suggestions"   default(5)
// @Success      200 {array} string
// @Router       /public/search/suggest [get]
func (h *Handler) PublicSuggest(c *gin.Context) {
	prefix := c.Query("q")
	if prefix == "" {
		c.JSON(http.StatusOK, []string{})
		return
	}
	locale := c.DefaultQuery("locale", "zh")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	suggestions, err := h.search.Suggest(c.Request.Context(), prefix, locale, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggest failed"})
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

// AdminRebuildIndex rebuilds the search index.
// @Summary      Rebuild search index
// @Description  Rebuild the full-text search index from scratch
// @Tags         Search (Admin)
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{message=string}
// @Router       /admin/search/rebuild [post]
func (h *Handler) AdminRebuildIndex(c *gin.Context) {
	if err := h.search.RebuildIndex(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rebuild failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "index rebuilt successfully"})
}

func (h *Handler) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/search", h.PublicSearch)
	public.GET("/search/suggest", h.PublicSuggest)
	admin.POST("/search/rebuild", h.AdminRebuildIndex)
}
