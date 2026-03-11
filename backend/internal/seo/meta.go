package seo

import "html/template"

// PageMeta holds all meta tag values for server-side injection into index.html.
type PageMeta struct {
	Title        string
	Description  string
	Keywords     string
	CanonicalURL string
	Locale       string

	// Open Graph
	OgTitle       string
	OgDescription string
	OgImage       string
	OgURL         string
	OgType        string

	// Twitter Card
	TwitterCard string

	// JSON-LD script content (not escaped by template engine)
	JSONLD template.HTML
}

const (
	defaultTitle       = "印迹法规咨询 - 企业内设型法规团队 | 专业法规咨询服务"
	defaultDescription = "印迹法规咨询（Blotting Consultancy）- 为企业提供专业的内设型法规团队服务"
)

// DefaultPageMeta returns meta with sensible defaults matching current index.html.
func DefaultPageMeta() PageMeta {
	return PageMeta{
		Title:         defaultTitle,
		Description:   defaultDescription,
		Locale:        "zh",
		OgType:        "website",
		OgTitle:       defaultTitle,
		OgDescription: defaultDescription,
		TwitterCard:   "summary_large_image",
	}
}
