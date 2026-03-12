package form_submission

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// Handler handles form submission HTTP requests
type Handler struct {
	repo repository.FormSubmissionRepository
}

// NewHandler creates a new form submission handler
func NewHandler(repo repository.FormSubmissionRepository) *Handler {
	return &Handler{repo: repo}
}

// --- Public endpoint ---

// submitInput is the JSON body for public form submission
type submitInput struct {
	FormType  string        `json:"formType"`
	Name      string        `json:"name"`
	Email     string        `json:"email"`
	Phone     string        `json:"phone"`
	Company   string        `json:"company"`
	Message   string        `json:"message"`
	SourceURL string        `json:"sourceUrl"`
	Locale    string        `json:"locale"`
	Metadata  model.JSONMap `json:"metadata"`
}

// HandlePublicSubmit handles a public form submission.
// @Summary      Submit form
// @Description  Submit a public contact/inquiry form
// @Tags         Form Submissions
// @Accept       json
// @Produce      json
// @Param        body body object true "Form submission data"
// @Success      201 {object} object
// @Failure      400 {object} object{error=string}
// @Router       /public/form-submissions [post]
func (h *Handler) HandlePublicSubmit(c *gin.Context) {
	var input submitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request data"}})
		return
	}

	submission := &model.FormSubmission{
		FormType:  input.FormType,
		Name:      input.Name,
		Email:     input.Email,
		Phone:     input.Phone,
		Company:   input.Company,
		Message:   input.Message,
		SourceURL: input.SourceURL,
		Locale:    input.Locale,
		IPAddress: c.ClientIP(),
		Status:    model.SubmissionStatusUnread,
		Metadata:  input.Metadata,
	}

	if err := h.repo.Create(c.Request.Context(), submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, submission)
}

// --- Admin endpoints ---

// HandleAdminList returns paginated form submissions.
// @Summary      List form submissions (admin)
// @Description  Returns paginated form submissions with optional filters
// @Tags         Form Submissions (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        page     query int    false "Page number"    default(1)
// @Param        pageSize query int    false "Items per page" default(20)
// @Param        formType query string false "Form type filter"
// @Param        status   query string false "Status filter (unread/read/archived)"
// @Success      200 {object} object{items=[]object,total=int,page=int,pageSize=int}
// @Router       /admin/form-submissions [get]
func (h *Handler) HandleAdminList(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	formType := c.Query("formType")
	status := c.Query("status")

	items, total, err := h.repo.List(c.Request.Context(), offset, pageSize, formType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "query failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// HandleAdminCounts returns submission counts grouped by status.
// @Summary      Get submission counts
// @Description  Returns submission counts grouped by status
// @Tags         Form Submissions (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        formType query string false "Form type filter"
// @Success      200 {object} object{counts=object}
// @Router       /admin/form-submissions/counts [get]
func (h *Handler) HandleAdminCounts(c *gin.Context) {
	formType := c.Query("formType")

	counts, err := h.repo.CountByStatus(c.Request.Context(), formType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "query failed"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"counts": counts})
}

// HandleAdminGetByID returns a single form submission by ID.
// @Summary      Get form submission by ID
// @Description  Returns a single form submission
// @Tags         Form Submissions (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Submission ID"
// @Success      200 {object} object
// @Failure      404 {object} object{error=string}
// @Router       /admin/form-submissions/{id} [get]
func (h *Handler) HandleAdminGetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid ID"}})
		return
	}

	submission, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "submission not found"}})
		return
	}

	c.JSON(http.StatusOK, submission)
}

// statusUpdateInput is the JSON body for updating submission status
type statusUpdateInput struct {
	Status string `json:"status"`
}

// HandleAdminUpdateStatus updates the status of a form submission.
// @Summary      Update submission status
// @Description  Update the status of a single form submission
// @Tags         Form Submissions (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int    true "Submission ID"
// @Param        body body object true "Status update"
// @Success      200 {object} object
// @Failure      404 {object} object{error=string}
// @Router       /admin/form-submissions/{id}/status [patch]
func (h *Handler) HandleAdminUpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid ID"}})
		return
	}

	submission, err := h.repo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "submission not found"}})
		return
	}

	var input statusUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request data"}})
		return
	}

	submission.Status = model.SubmissionStatus(input.Status)

	if err := h.repo.Update(c.Request.Context(), submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, submission)
}

// bulkStatusInput is the JSON body for bulk status update
type bulkStatusInput struct {
	IDs    []uint `json:"ids"`
	Status string `json:"status"`
}

// HandleAdminBulkUpdateStatus updates status of multiple submissions.
// @Summary      Bulk update status
// @Description  Update the status of multiple form submissions at once
// @Tags         Form Submissions (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body object true "IDs and target status"
// @Success      200 {object} object{message=string,count=int}
// @Router       /admin/form-submissions/bulk-status [post]
func (h *Handler) HandleAdminBulkUpdateStatus(c *gin.Context) {
	var input bulkStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request data"}})
		return
	}

	status := model.SubmissionStatus(input.Status)
	if status != model.SubmissionStatusUnread && status != model.SubmissionStatusRead && status != model.SubmissionStatusArchived {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "status must be unread, read, or archived"}})
		return
	}

	if err := h.repo.BulkUpdateStatus(c.Request.Context(), input.IDs, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已更新",
		"count":   len(input.IDs),
	})
}

// HandleAdminDelete soft-deletes a form submission.
// @Summary      Delete form submission
// @Description  Soft-delete a form submission by ID
// @Tags         Form Submissions (Admin)
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Submission ID"
// @Success      200 {object} object{message=string}
// @Failure      404 {object} object{error=string}
// @Router       /admin/form-submissions/{id} [delete]
func (h *Handler) HandleAdminDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid ID"}})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "submission not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
