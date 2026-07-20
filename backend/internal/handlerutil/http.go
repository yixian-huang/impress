package handlerutil

import (
	"strconv"

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
