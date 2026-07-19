package feed

import (
	"encoding/xml"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

// Handler serves RSS 2.0 feeds for published articles.
type Handler struct {
	articleRepo  repository.ArticleRepository
	siteCfgRepo  repository.SiteConfigRepository
	baseURL      string
	channelTitle string
	channelDesc  string
}

func NewHandler(
	articleRepo repository.ArticleRepository,
	siteCfgRepo repository.SiteConfigRepository,
	baseURL string,
	channelTitle string,
	channelDesc string,
) *Handler {
	return &Handler{
		articleRepo:  articleRepo,
		siteCfgRepo:  siteCfgRepo,
		baseURL:      strings.TrimRight(baseURL, "/"),
		channelTitle: channelTitle,
		channelDesc:  channelDesc,
	}
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language,omitempty"`
	LastBuild   string    `xml:"lastBuildDate,omitempty"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func (h *Handler) rssEnabled(ctx *gin.Context) bool {
	if h.siteCfgRepo == nil {
		return true
	}
	cfg, err := h.siteCfgRepo.FindByKey(ctx.Request.Context(), model.SiteConfigKeyFeatures)
	if err != nil || cfg == nil || cfg.ID == 0 || cfg.PublishedConfig == nil {
		return true
	}
	blog, ok := cfg.PublishedConfig["blog"].(map[string]interface{})
	if !ok {
		return true
	}
	rss, ok := blog["rss"].(bool)
	if !ok {
		return true
	}
	return rss
}

func stripHTML(s string) string {
	out := s
	for {
		i := strings.Index(out, "<")
		if i < 0 {
			break
		}
		j := strings.Index(out[i:], ">")
		if j < 0 {
			break
		}
		out = out[:i] + out[i+j+1:]
	}
	return strings.TrimSpace(html.UnescapeString(out))
}

func itemDescription(article *model.Article) string {
	body := article.ZhBody
	if body == "" {
		body = article.EnBody
	}
	if article.AutoSummary || body == "" {
		desc := article.ZhMetaDescription
		if desc == "" {
			desc = article.EnMetaDescription
		}
		if desc != "" {
			return desc
		}
	}
	text := stripHTML(body)
	if len(text) > 500 {
		return text[:500] + "..."
	}
	return text
}

func pubDate(t *time.Time, fallback time.Time) string {
	if t != nil && !t.IsZero() {
		return t.UTC().Format(time.RFC1123Z)
	}
	return fallback.UTC().Format(time.RFC1123Z)
}

// GetFeed returns RSS 2.0 XML for published articles.
func (h *Handler) GetFeed(c *gin.Context) {
	if !h.rssEnabled(c) {
		c.Status(http.StatusNotFound)
		return
	}

	limit := 50
	articles, _, err := h.articleRepo.ListPublished(c.Request.Context(), 0, limit, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "feed generation failed"}})
		return
	}

	items := make([]rssItem, 0, len(articles))
	for _, a := range articles {
		if a.Visibility != "" && a.Visibility != "public" {
			continue
		}
		title := a.ZhTitle
		if title == "" {
			title = a.EnTitle
		}
		link := h.baseURL + "/blog/" + a.Slug
		items = append(items, rssItem{
			Title:       title,
			Link:        link,
			Description: itemDescription(a),
			GUID:        link,
			PubDate:     pubDate(a.PublishedAt, a.CreatedAt),
		})
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       h.channelTitle,
			Link:        h.baseURL + "/",
			Description: h.channelDesc,
			Language:    "zh-cn",
			LastBuild:   time.Now().UTC().Format(time.RFC1123Z),
			Items:       items,
		},
	}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=300")
	c.XML(http.StatusOK, feed)
}
