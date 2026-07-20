package apikey

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/middleware"
	"github.com/yixian-huang/inkless/backend/internal/service"
	"github.com/yixian-huang/inkless/backend/pkg/apierror"
)

// Handler manages personal API keys for the authenticated user.
type Handler struct {
	svc *service.APIKeyService
}

func NewHandler(svc *service.APIKeyService) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// List GET /admin/api-keys
func (h *Handler) List(c *gin.Context) {
	uc := middleware.GetUserContext(c)
	if uc == nil {
		apierror.Message(c, http.StatusUnauthorized, "需要登录")
		return
	}
	list, err := h.svc.List(c.Request.Context(), uc.UserID)
	if err != nil {
		apierror.Message(c, http.StatusInternalServerError, "列出 API Key 失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Create POST /admin/api-keys — response includes token plaintext once.
func (h *Handler) Create(c *gin.Context) {
	uc := middleware.GetUserContext(c)
	if uc == nil {
		apierror.Message(c, http.StatusUnauthorized, "需要登录")
		return
	}
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Message(c, http.StatusBadRequest, "请求体无效")
		return
	}
	plain, key, err := h.svc.Create(c.Request.Context(), uc.UserID, req.Name, req.Scopes)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAPIKeyInvalidName),
			errors.Is(err, service.ErrAPIKeyInvalidScope),
			errors.Is(err, service.ErrAPIKeyLimit):
			apierror.Message(c, http.StatusBadRequest, err.Error())
		default:
			apierror.Message(c, http.StatusInternalServerError, "创建 API Key 失败")
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token": plain, // only once
		"key":   key,
	})
}

// Revoke DELETE /admin/api-keys/:id
func (h *Handler) Revoke(c *gin.Context) {
	uc := middleware.GetUserContext(c)
	if uc == nil {
		apierror.Message(c, http.StatusUnauthorized, "需要登录")
		return
	}
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		apierror.Message(c, http.StatusBadRequest, "无效 id")
		return
	}
	if err := h.svc.Revoke(c.Request.Context(), uc.UserID, uint(id64)); err != nil {
		if errors.Is(err, service.ErrAPIKeyNotFound) {
			apierror.Message(c, http.StatusNotFound, "API Key 不存在")
			return
		}
		apierror.Message(c, http.StatusInternalServerError, "吊销失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已吊销"})
}
