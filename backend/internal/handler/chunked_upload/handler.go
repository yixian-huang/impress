package chunked_upload

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/apierror"

	"github.com/yixian-huang/inkless/backend/internal/service"
)

// Handler handles chunked upload HTTP requests
type Handler struct {
	svc *service.ChunkedUploadService
}

// NewHandler creates a new chunked upload handler
func NewHandler(svc *service.ChunkedUploadService) *Handler {
	return &Handler{svc: svc}
}

// InitUpload initializes a new chunked upload
// POST /admin/media/upload/init
func (h *Handler) InitUpload(c *gin.Context) {
	var req service.InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Message(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	upload, err := h.svc.InitUpload(c.Request.Context(), &req)
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"uploadId":    upload.ID,
		"totalChunks": upload.TotalChunks,
		"chunkSize":   upload.ChunkSize,
		"expiresAt":   upload.ExpiresAt,
	})
}

// UploadChunk uploads a single chunk
// POST /admin/media/upload/:uploadId/chunk
func (h *Handler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("uploadId")
	if uploadID == "" {
		apierror.Message(c, http.StatusBadRequest, "缺少 uploadId")
		return
	}

	chunkIndexStr := c.PostForm("chunkIndex")
	if chunkIndexStr == "" {
		chunkIndexStr = c.Query("chunkIndex")
	}
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, "无效的 chunkIndex")
		return
	}

	file, _, err := c.Request.FormFile("chunk")
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, "缺少 chunk 文件")
		return
	}
	defer file.Close()

	if err := h.svc.UploadChunk(c.Request.Context(), uploadID, chunkIndex, file); err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "分片上传成功",
		"chunkIndex": chunkIndex,
	})
}

// CompleteUpload finalizes the chunked upload and merges chunks
// POST /admin/media/upload/:uploadId/complete
func (h *Handler) CompleteUpload(c *gin.Context) {
	uploadID := c.Param("uploadId")
	if uploadID == "" {
		apierror.Message(c, http.StatusBadRequest, "缺少 uploadId")
		return
	}

	media, err := h.svc.CompleteUpload(c.Request.Context(), uploadID)
	if err != nil {
		apierror.Message(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, media)
}
