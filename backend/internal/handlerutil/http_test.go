package handlerutil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/handlerutil"
)

func TestParsePaginationClamps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=0&pageSize=999", nil)

	p := handlerutil.ParsePagination(c, 20, 50)
	require.Equal(t, 1, p.Page)
	require.Equal(t, 50, p.PageSize)
	require.Equal(t, 0, p.Offset)
}

func TestParseUintParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/items/:id", func(c *gin.Context) {
		id, ok := handlerutil.ParseUintParam(c, "id")
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/items/abc", nil)
	r.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusBadRequest, w2.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &body))
	_, hasError := body["error"]
	require.True(t, hasError)
	_ = c
}
