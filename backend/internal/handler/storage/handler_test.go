package storage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDecodeUpdateConfigRequestRejectsNonCamelCaseFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodPut, "/storage", strings.NewReader(`{
		"strategy":"s3",
		"access_key":"ak",
		"secretKey":"sk"
	}`))
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	if _, err := decodeUpdateConfigRequest(ctx); err == nil {
		t.Fatal("expected non-camelCase field to be rejected")
	}
}
