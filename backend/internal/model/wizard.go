package model

// Questionnaire holds the user's answers from the site building wizard.
type Questionnaire struct {
	// Industry is the user's industry/business type (e.g., "technology", "consulting", "retail").
	Industry string `json:"industry" binding:"required"`

	// StylePreference is the visual style preference (e.g., "modern", "classic", "minimalist", "bold").
	StylePreference string `json:"stylePreference"`

	// Features is a list of desired functionality (e.g., ["blog", "contact_form", "portfolio"]).
	Features []string `json:"features"`

	// ContentTypes describes what kind of content the site will feature
	// (e.g., ["articles", "case_studies", "team_profiles"]).
	ContentTypes []string `json:"contentTypes"`

	// BrandName is the name of the brand or company.
	BrandName string `json:"brandName"`

	// Description is a short description of the business or site purpose.
	Description string `json:"description"`

	// Locale is the primary locale for generated content ("zh" or "en"). Defaults to "zh".
	Locale string `json:"locale"`
}

// PagePlan describes a single page to be scaffolded.
type PagePlan struct {
	// Slug is the URL slug for the page.
	Slug string `json:"slug"`

	// Title holds bilingual titles (keyed by locale, e.g., "zh"/"en").
	Title map[string]string `json:"title"`

	// Layout is the recommended layout type (e.g., "hero-cta", "grid", "split").
	Layout string `json:"layout"`

	// Sections lists the recommended section types (e.g., ["hero", "features", "contact"]).
	Sections []string `json:"sections"`

	// SortOrder is the display order of the page in navigation.
	SortOrder int `json:"sortOrder"`
}

// ColorScheme holds a recommended brand color palette.
type ColorScheme struct {
	// Primary is the main brand color in hex (e.g., "#2563EB").
	Primary string `json:"primary"`

	// Secondary is the accent color in hex.
	Secondary string `json:"secondary"`

	// Background is the recommended page background color.
	Background string `json:"background"`

	// Text is the recommended body text color.
	Text string `json:"text"`

	// Rationale explains why this palette was chosen.
	Rationale string `json:"rationale"`
}

// SuggestedContent holds AI-generated sample content for a specific page.
type SuggestedContent struct {
	// PageSlug is the target page slug this content is intended for.
	PageSlug string `json:"pageSlug"`

	// Heading is the main headline.
	Heading string `json:"heading"`

	// Subheading is the supporting subtitle or tagline.
	Subheading string `json:"subheading"`

	// Body is the main body copy.
	Body string `json:"body"`

	// CTAText is the suggested call-to-action button text.
	CTAText string `json:"ctaText"`
}

// SitePlan is the AI-generated site structure recommendation.
type SitePlan struct {
	// RecommendedTheme is the theme package name (e.g., "default", "modern-dark", "warm-earth").
	RecommendedTheme string `json:"recommendedTheme"`

	// Pages is the list of pages to scaffold.
	Pages []PagePlan `json:"pages"`

	// ColorScheme is the recommended color palette.
	ColorScheme ColorScheme `json:"colorScheme"`

	// SuggestedContent contains AI-generated sample content per page.
	SuggestedContent []SuggestedContent `json:"suggestedContent"`

	// Rationale explains the overall design and structure decisions.
	Rationale string `json:"rationale"`
}

// ScaffoldResult holds the outcome of applying a site plan.
type ScaffoldResult struct {
	// CreatedPages lists the slugs of pages successfully created.
	CreatedPages []string `json:"createdPages"`

	// SkippedPages lists slugs that were skipped because they already existed.
	SkippedPages []string `json:"skippedPages"`

	// AppliedTheme is the theme that was recommended (theme activation is handled separately).
	AppliedTheme string `json:"appliedTheme"`
}

// ColorSuggestionRequest is the input for the suggest-colors endpoint.
type ColorSuggestionRequest struct {
	Industry  string `json:"industry" binding:"required"`
	BrandName string `json:"brandName"`
	Locale    string `json:"locale"`
}

// GenerateContentRequest is the input for the generate-content endpoint.
type GenerateContentRequest struct {
	PageType    string `json:"pageType" binding:"required"`
	Industry    string `json:"industry" binding:"required"`
	BrandName   string `json:"brandName"`
	Description string `json:"description"`
	Locale      string `json:"locale"`
}
