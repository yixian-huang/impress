package article

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/model"
)

// AdminExportMarkdown handles GET /admin/articles/:id/export
// Returns the article as a Markdown file with YAML front matter.
func (h *Handler) AdminExportMarkdown(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "无效的 ID"}})
		return
	}

	article, err := h.articleRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "文章不存在"}})
		return
	}

	// Build YAML front matter + body
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %q\n", article.ZhTitle))
	if article.EnTitle != "" {
		sb.WriteString(fmt.Sprintf("title_en: %q\n", article.EnTitle))
	}
	sb.WriteString(fmt.Sprintf("slug: %s\n", article.Slug))
	sb.WriteString(fmt.Sprintf("status: %s\n", article.Status))
	if article.PublishedAt != nil {
		sb.WriteString(fmt.Sprintf("date: %s\n", article.PublishedAt.Format("2006-01-02")))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(article.ZhBody)

	filename := article.Slug + ".md"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(sb.String()))
}

// AdminImportMarkdown handles POST /admin/articles/import
// Accepts multipart form with .md files. Creates draft articles.
func (h *Handler) AdminImportMarkdown(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "multipart form required"}})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "no files provided"}})
		return
	}

	var imported []string
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}

		title, body := parseMarkdownFrontMatter(string(content))
		if title == "" {
			title = strings.TrimSuffix(file.Filename, ".md")
		}

		slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))

		article := &model.Article{
			ZhTitle: title,
			ZhBody:  body,
			Slug:    slug,
			Status:  model.ArticleStatusDraft,
		}

		if err := h.articleRepo.Create(c.Request.Context(), article); err != nil {
			continue
		}
		imported = append(imported, title)
	}

	c.JSON(http.StatusOK, gin.H{"imported": imported, "count": len(imported)})
}

// parseMarkdownFrontMatter extracts title from YAML front matter and returns body.
func parseMarkdownFrontMatter(content string) (title, body string) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content
	}

	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return "", content
	}

	frontMatter := content[4 : 4+end]
	body = strings.TrimSpace(content[4+end+4:])

	for _, line := range strings.Split(frontMatter, "\n") {
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			title = strings.Trim(title, "\"'")
		}
	}

	return title, body
}
