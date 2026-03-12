package translation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
)

// Handler handles translation and glossary HTTP requests
type Handler struct {
	translator   provider.TranslationProvider
	glossaryRepo repository.GlossaryRepository
	articleRepo  repository.ArticleRepository
}

// NewHandler creates a new translation handler
func NewHandler(
	translator provider.TranslationProvider,
	glossaryRepo repository.GlossaryRepository,
	articleRepo repository.ArticleRepository,
) *Handler {
	return &Handler{
		translator:   translator,
		glossaryRepo: glossaryRepo,
		articleRepo:  articleRepo,
	}
}

// --- Translation endpoints ---

// translateInput is the JSON body for a translation request
type translateInput struct {
	Text       string `json:"text" binding:"required"`
	SourceLang string `json:"sourceLang" binding:"required"`
	TargetLang string `json:"targetLang" binding:"required"`
}

// Translate handles a single text translation
// POST /admin/translate
func (h *Handler) Translate(c *gin.Context) {
	var input translateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request: text, sourceLang, and targetLang are required"}})
		return
	}

	// Load glossary for the language pair
	glossary, err := h.buildGlossaryMap(c, input.SourceLang, input.TargetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to load glossary"}})
		return
	}

	req := provider.TranslateRequest{
		Text:       input.Text,
		SourceLang: input.SourceLang,
		TargetLang: input.TargetLang,
		Glossary:   glossary,
	}

	resp, err := h.translator.Translate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// batchTranslateInput is the JSON body for a batch translation request
type batchTranslateInput struct {
	Items []translateInput `json:"items" binding:"required"`
}

// BatchTranslate handles batch text translations
// POST /admin/translate/batch
func (h *Handler) BatchTranslate(c *gin.Context) {
	var input batchTranslateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request: items array is required"}})
		return
	}

	if len(input.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "items array must not be empty"}})
		return
	}

	reqs := make([]provider.TranslateRequest, 0, len(input.Items))
	for _, item := range input.Items {
		glossary, err := h.buildGlossaryMap(c, item.SourceLang, item.TargetLang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to load glossary"}})
			return
		}
		reqs = append(reqs, provider.TranslateRequest{
			Text:       item.Text,
			SourceLang: item.SourceLang,
			TargetLang: item.TargetLang,
			Glossary:   glossary,
		})
	}

	responses, err := h.translator.BatchTranslate(c.Request.Context(), reqs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"translations": responses})
}

// TranslateArticle translates an article's content fields
// POST /admin/translate/article/:id
func (h *Handler) TranslateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid article ID"}})
		return
	}

	article, err := h.articleRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "article not found"}})
		return
	}

	// Determine source and target: if Chinese content exists, translate zh->en; otherwise en->zh
	sourceLang := "zh"
	targetLang := "en"
	if article.ZhTitle == "" && article.EnTitle != "" {
		sourceLang = "en"
		targetLang = "zh"
	}

	// Load glossary
	glossary, err := h.buildGlossaryMap(c, sourceLang, targetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to load glossary"}})
		return
	}

	// Build translation requests for non-empty source fields
	type fieldMapping struct {
		sourceText string
		fieldName  string
	}

	var mappings []fieldMapping
	if sourceLang == "zh" {
		if article.ZhTitle != "" {
			mappings = append(mappings, fieldMapping{sourceText: article.ZhTitle, fieldName: "title"})
		}
		if article.ZhBody != "" {
			mappings = append(mappings, fieldMapping{sourceText: article.ZhBody, fieldName: "body"})
		}
	} else {
		if article.EnTitle != "" {
			mappings = append(mappings, fieldMapping{sourceText: article.EnTitle, fieldName: "title"})
		}
		if article.EnBody != "" {
			mappings = append(mappings, fieldMapping{sourceText: article.EnBody, fieldName: "body"})
		}
	}

	if len(mappings) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "no source content to translate"}})
		return
	}

	reqs := make([]provider.TranslateRequest, len(mappings))
	for i, m := range mappings {
		reqs[i] = provider.TranslateRequest{
			Text:       m.sourceText,
			SourceLang: sourceLang,
			TargetLang: targetLang,
			Glossary:   glossary,
		}
	}

	responses, err := h.translator.BatchTranslate(c.Request.Context(), reqs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	// Apply translations to the article
	translations := make(map[string]string)
	for i, m := range mappings {
		translations[m.fieldName] = responses[i].TranslatedText
	}

	if targetLang == "en" {
		if v, ok := translations["title"]; ok {
			article.EnTitle = v
		}
		if v, ok := translations["body"]; ok {
			article.EnBody = v
		}
	} else {
		if v, ok := translations["title"]; ok {
			article.ZhTitle = v
		}
		if v, ok := translations["body"]; ok {
			article.ZhBody = v
		}
	}

	if err := h.articleRepo.Update(c.Request.Context(), article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "failed to save translated article"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"article":      article,
		"sourceLang":   sourceLang,
		"targetLang":   targetLang,
		"translations": translations,
	})
}

// --- Glossary endpoints ---

// GlossaryList returns a paginated list of glossary terms
// GET /admin/glossary?page=1&pageSize=20&sourceLang=zh&targetLang=en
func (h *Handler) GlossaryList(c *gin.Context) {
	page := 1
	pageSize := 50

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
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize
	sourceLang := c.Query("sourceLang")
	targetLang := c.Query("targetLang")

	items, total, err := h.glossaryRepo.List(c.Request.Context(), offset, pageSize, sourceLang, targetLang)
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

// glossaryInput is the JSON body for creating/updating a glossary term
type glossaryInput struct {
	SourceLang string `json:"sourceLang" binding:"required"`
	TargetLang string `json:"targetLang" binding:"required"`
	SourceTerm string `json:"sourceTerm" binding:"required"`
	TargetTerm string `json:"targetTerm" binding:"required"`
	Context    string `json:"context"`
}

// GlossaryCreate creates a new glossary term
// POST /admin/glossary
func (h *Handler) GlossaryCreate(c *gin.Context) {
	var input glossaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request: sourceLang, targetLang, sourceTerm, and targetTerm are required"}})
		return
	}

	glossary := &model.Glossary{
		SourceLang: input.SourceLang,
		TargetLang: input.TargetLang,
		SourceTerm: input.SourceTerm,
		TargetTerm: input.TargetTerm,
		Context:    input.Context,
	}

	if err := h.glossaryRepo.Create(c.Request.Context(), glossary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, glossary)
}

// GlossaryUpdate updates a glossary term
// PUT /admin/glossary/:id
func (h *Handler) GlossaryUpdate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid ID"}})
		return
	}

	existing, err := h.glossaryRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "glossary term not found"}})
		return
	}

	var input glossaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid request data"}})
		return
	}

	existing.SourceLang = input.SourceLang
	existing.TargetLang = input.TargetLang
	existing.SourceTerm = input.SourceTerm
	existing.TargetTerm = input.TargetTerm
	existing.Context = input.Context

	if err := h.glossaryRepo.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// GlossaryDelete deletes a glossary term
// DELETE /admin/glossary/:id
func (h *Handler) GlossaryDelete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "invalid ID"}})
		return
	}

	if err := h.glossaryRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "glossary term not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// buildGlossaryMap loads glossary terms for a language pair and returns them as a map
func (h *Handler) buildGlossaryMap(c *gin.Context, sourceLang, targetLang string) (map[string]string, error) {
	terms, err := h.glossaryRepo.FindByLangs(c.Request.Context(), sourceLang, targetLang)
	if err != nil {
		return nil, err
	}

	glossary := make(map[string]string, len(terms))
	for _, term := range terms {
		glossary[term.SourceTerm] = term.TargetTerm
	}
	return glossary, nil
}
