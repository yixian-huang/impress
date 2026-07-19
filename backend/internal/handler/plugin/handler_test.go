package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestExternalPluginMutationIsDisabledByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, false)
	router := gin.New()
	router.POST("/plugins/install", handler.Install)

	request := httptest.NewRequest(http.MethodPost, "/plugins/install", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Contains(t, response.Body.String(), "ENABLE_EXTERNAL_PLUGINS=true")
}

func TestPluginResponseMasksSettingsAndBinaryPath(t *testing.T) {
	response := pluginResponse(model.Plugin{
		PluginID:   "test-plugin",
		BinaryPath: "/srv/inkless/plugins/test-plugin/plugin",
		Settings:   model.JSONMap{"outputFile": "events.jsonl"},
	})

	assert.NotContains(t, response, "settings")
	assert.NotContains(t, response, "binaryPath")
	assert.Equal(t, true, response["hasSettings"])
}
