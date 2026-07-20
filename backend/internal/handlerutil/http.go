package handlerutil

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/apierror"
)

// DefaultMaxPageSize is the clamp used by admin/public list endpoints.
const DefaultMaxPageSize = 100

// ParseUintParam parses a positive uint path/query param.
// On failure it writes a 400 response and returns ok=false.
func ParseUintParam(c *gin.Context, name string) (id uint, ok bool) {
	raw := c.Param(name)
	if raw == "" {
		raw = c.Query(name)
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		apierror.Write(c, apierror.BadRequest("invalid "+name))
		return 0, false
	}
	return uint(parsed), true
}

// ParseUintParamOptional returns 0 when the param/query is empty.
// A present but non-positive value is a 400.
func ParseUintParamOptional(c *gin.Context, name string) (id uint, ok bool) {
	raw := c.Param(name)
	if raw == "" {
		raw = c.Query(name)
	}
	if raw == "" {
		return 0, true
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		apierror.Write(c, apierror.BadRequest("invalid "+name))
		return 0, false
	}
	return uint(parsed), true
}

// Pagination holds offset/limit derived from page/pageSize query params.
type Pagination struct {
	Page     int
	PageSize int
	Offset   int
}

// ParsePagination reads page (default 1) and pageSize (default defaultSize, max maxSize).
func ParsePagination(c *gin.Context, defaultSize, maxSize int) Pagination {
	if defaultSize <= 0 {
		defaultSize = 20
	}
	if maxSize <= 0 {
		maxSize = DefaultMaxPageSize
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(defaultSize)))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// ListResponse is the standard admin list JSON shape.
func ListResponse(c *gin.Context, items any, total int64, p Pagination) {
	c.JSON(200, map[string]any{
		"items":    items,
		"total":    total,
		"page":     p.Page,
		"pageSize": p.PageSize,
	})
}

// QueryTrim returns a trimmed query string param (empty if blank).
func QueryTrim(c *gin.Context, name string) string {
	return strings.TrimSpace(c.Query(name))
}

// ParseSort allows only keys in allowed; returns defaultKey when empty/invalid.
// allowed maps request token → SQL ORDER BY clause (already sanitized).
func ParseSort(c *gin.Context, param string, allowed map[string]string, defaultKey string) string {
	raw := strings.TrimSpace(c.Query(param))
	if raw == "" {
		raw = defaultKey
	}
	if clause, ok := allowed[raw]; ok {
		return clause
	}
	if clause, ok := allowed[defaultKey]; ok {
		return clause
	}
	return "created_at DESC"
}
