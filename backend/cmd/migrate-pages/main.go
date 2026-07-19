package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/pkg/config"
)

// pageOrder defines the sort order for migrated pages, matching ValidPageKeys order.
var pageOrder = map[model.PageKey]int{
	model.PageKeyHome:         0,
	model.PageKeyAbout:        1,
	model.PageKeyAdvantages:   2,
	model.PageKeyCoreServices: 3,
	model.PageKeyCases:        4,
	model.PageKeyExperts:      5,
	model.PageKeyContact:      6,
}

// migrableKeys are the page keys that should be migrated (excluding global and theme).
var migrableKeys = []model.PageKey{
	model.PageKeyHome,
	model.PageKeyAbout,
	model.PageKeyAdvantages,
	model.PageKeyCoreServices,
	model.PageKeyCases,
	model.PageKeyExperts,
	model.PageKeyContact,
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	maxOpen := 25
	maxIdle := 5
	maxLife := 5 * time.Minute
	if !db.IsPostgresDSN(cfg.DBDSN) {
		maxOpen = 1
		maxIdle = 1
		maxLife = 0
	}

	database, err := db.Init(db.InitOptions{
		DSN:         cfg.DBDSN,
		MaxOpenConn: maxOpen,
		MaxIdleConn: maxIdle,
		MaxLifetime: maxLife,
		LogLevel:    logger.Warn,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Ensure pages table exists
	if err := database.AutoMigrate(&model.Page{}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to migrate pages table: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== ContentDocument → Page Migration ===")
	fmt.Println()

	var created, skipped int

	for _, key := range migrableKeys {
		slug := string(key)

		// Idempotency: skip if page with this slug already exists
		var existing model.Page
		if err := database.DB.Where("slug = ?", slug).First(&existing).Error; err == nil {
			fmt.Printf("  SKIP  %s → /%s (already exists, id=%d)\n", key, slug, existing.ID)
			skipped++
			continue
		}

		// Load the ContentDocument
		var doc model.ContentDocument
		if err := database.DB.Where("page_key = ?", string(key)).First(&doc).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("  SKIP  %s → no ContentDocument found\n", key)
				skipped++
				continue
			}
			fmt.Fprintf(os.Stderr, "  ERROR %s → %v\n", key, err)
			os.Exit(1)
		}

		cfg := doc.PublishedConfig
		if len(cfg) == 0 {
			cfg = doc.DraftConfig
		}
		if len(cfg) == 0 {
			fmt.Printf("  SKIP  %s → empty config\n", key)
			skipped++
			continue
		}

		sections := convertToSections(key, cfg)

		title := extractTitle(key, cfg)

		pageConfig := model.JSONMap{
			"sections": sections,
		}

		page := model.Page{
			Slug:      slug,
			Title:     title,
			Template:  "default",
			Config:    pageConfig,
			Status:    model.PageStatusPublished,
			SortOrder: pageOrder[key],
		}

		if err := database.DB.Create(&page).Error; err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR %s → failed to create page: %v\n", key, err)
			os.Exit(1)
		}

		fmt.Printf("  OK    %s → /%s (%d sections, id=%d)\n", key, slug, len(sections), page.ID)
		created++
	}

	fmt.Println()
	fmt.Printf("Done: %d created, %d skipped\n", created, skipped)
}

// section is a helper to build a section map.
func section(id, sectionType string, data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":   id,
		"type": sectionType,
		"data": data,
	}
}

// getStr safely extracts a string from a nested config path.
func getStr(m model.JSONMap, keys ...string) string {
	var current interface{} = map[string]interface{}(m)
	for _, k := range keys {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = obj[k]
	}
	if s, ok := current.(string); ok {
		return s
	}
	return ""
}

// getMap safely extracts a sub-map from a config.
func getMap(m model.JSONMap, keys ...string) map[string]interface{} {
	var current interface{} = map[string]interface{}(m)
	for _, k := range keys {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = obj[k]
	}
	if obj, ok := current.(map[string]interface{}); ok {
		return obj
	}
	return nil
}

// getSlice safely extracts a slice from a config.
func getSlice(m model.JSONMap, keys ...string) []interface{} {
	var current interface{} = map[string]interface{}(m)
	for _, k := range keys {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = obj[k]
	}
	if arr, ok := current.([]interface{}); ok {
		return arr
	}
	return nil
}

// extractTitle extracts a bilingual title from the config.
func extractTitle(key model.PageKey, cfg model.JSONMap) model.JSONMap {
	title := model.JSONMap{}

	// Try hero title or top-level title
	if zh := getStr(cfg, "title"); zh != "" {
		title["zh"] = zh
	} else if zh := getStr(cfg, "hero", "title"); zh != "" {
		title["zh"] = zh
	}

	if en := getStr(cfg, "titleEn"); en != "" {
		title["en"] = en
	} else if en := getStr(cfg, "hero", "titleEn"); en != "" {
		title["en"] = en
	}

	// Fallback to page key
	if len(title) == 0 {
		title["zh"] = string(key)
		title["en"] = string(key)
	}

	return title
}

func convertToSections(key model.PageKey, cfg model.JSONMap) []map[string]interface{} {
	switch key {
	case model.PageKeyHome:
		return convertHome(cfg)
	case model.PageKeyAbout:
		return convertAbout(cfg)
	case model.PageKeyAdvantages:
		return convertAdvantages(cfg)
	case model.PageKeyCoreServices:
		return convertCoreServices(cfg)
	case model.PageKeyExperts:
		return convertExperts(cfg)
	case model.PageKeyCases:
		return convertCases(cfg)
	case model.PageKeyContact:
		return convertContact(cfg)
	default:
		return nil
	}
}

func convertHome(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero section
	heroData := map[string]interface{}{}
	copyIfPresent(heroData, cfg, "title", "subtitle", "backgroundImage")
	sections = append(sections, section("home-0", "hero", heroData))

	// Company profile section
	profileData := map[string]interface{}{}
	copyIfPresent(profileData, cfg, "title", "description", "description2", "description3", "button", "image")
	// Also check nested "companyProfile" key
	if cp := getMap(cfg, "companyProfile"); cp != nil {
		copyFromMap(profileData, cp, "title", "description", "description2", "description3", "button", "image")
	}
	sections = append(sections, section("home-1", "company-profile", profileData))

	// Card grid section
	cardGridData := map[string]interface{}{}
	if cg := getMap(cfg, "cardGrid"); cg != nil {
		copyFromMap(cardGridData, cg, "title")
		if cards := cg["cards"]; cards != nil {
			cardGridData["cards"] = cards
		}
	} else if cards := cfg["cards"]; cards != nil {
		cardGridData["cards"] = cards
		if t := cfg["cardsTitle"]; t != nil {
			cardGridData["title"] = t
		}
	}
	sections = append(sections, section("home-2", "card-grid", cardGridData))

	// Service cards section
	serviceData := map[string]interface{}{}
	if sc := getMap(cfg, "serviceCards"); sc != nil {
		copyFromMap(serviceData, sc, "title")
		if services := sc["services"]; services != nil {
			serviceData["services"] = services
		}
	} else if services := cfg["services"]; services != nil {
		serviceData["services"] = services
		if t := cfg["servicesTitle"]; t != nil {
			serviceData["title"] = t
		}
	}
	sections = append(sections, section("home-3", "service-cards", serviceData))

	return sections
}

func convertAbout(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "label", "title")
	} else {
		copyIfPresent(heroData, cfg, "label", "title")
	}
	sections = append(sections, section("about-0", "hero", heroData))

	// Company profile
	profileData := map[string]interface{}{}
	if cp := getMap(cfg, "companyProfile"); cp != nil {
		copyFromMap(profileData, cp, "title", "description")
	}
	sections = append(sections, section("about-1", "company-profile", profileData))

	// Text-image blocks
	blocks := getSlice(cfg, "blocks")
	if blocks == nil {
		blocks = getSlice(cfg, "textImageBlocks")
	}
	for i, b := range blocks {
		block, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		data := map[string]interface{}{}
		copyFromMap(data, block, "description", "image")
		if i%2 == 0 {
			data["imagePosition"] = "left"
		} else {
			data["imagePosition"] = "right"
		}
		sections = append(sections, section(fmt.Sprintf("about-%d", 2+i), "text-image", data))
	}

	return sections
}

func convertAdvantages(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "label", "title")
	} else {
		copyIfPresent(heroData, cfg, "label", "title")
	}
	sections = append(sections, section("advantages-0", "hero", heroData))

	// Advantage blocks
	blocks := getSlice(cfg, "blocks")
	if blocks == nil {
		blocks = getSlice(cfg, "advantages")
	}
	for i, b := range blocks {
		block, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		data := map[string]interface{}{}
		copyFromMap(data, block, "title", "description", "image")
		if i%2 == 0 {
			data["imagePosition"] = "left"
		} else {
			data["imagePosition"] = "right"
		}
		sections = append(sections, section(fmt.Sprintf("advantages-%d", 1+i), "text-image", data))
	}

	return sections
}

func convertCoreServices(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "label", "title")
	} else {
		copyIfPresent(heroData, cfg, "label", "title")
	}
	sections = append(sections, section("core-services-0", "hero", heroData))

	// Service blocks
	blocks := getSlice(cfg, "services")
	if blocks == nil {
		blocks = getSlice(cfg, "blocks")
	}
	for i, b := range blocks {
		block, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		data := map[string]interface{}{}
		copyFromMap(data, block, "title", "description", "image")
		if i%2 == 0 {
			data["imagePosition"] = "left"
		} else {
			data["imagePosition"] = "right"
		}
		sections = append(sections, section(fmt.Sprintf("core-services-%d", 1+i), "text-image", data))
	}

	return sections
}

func convertExperts(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "label", "title")
	} else {
		copyIfPresent(heroData, cfg, "label", "title")
	}
	sections = append(sections, section("experts-0", "hero", heroData))

	// Team grid
	teamData := map[string]interface{}{}
	if tg := getMap(cfg, "teamGrid"); tg != nil {
		copyFromMap(teamData, tg, "sectionTitle")
		if experts := tg["experts"]; experts != nil {
			teamData["experts"] = experts
		}
	} else {
		if st := cfg["sectionTitle"]; st != nil {
			teamData["sectionTitle"] = st
		}
		if experts := cfg["experts"]; experts != nil {
			teamData["experts"] = experts
		}
	}
	sections = append(sections, section("experts-1", "team-grid", teamData))

	return sections
}

func convertCases(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "label", "title", "imageSrc")
	} else {
		copyIfPresent(heroData, cfg, "label", "title", "imageSrc")
	}
	sections = append(sections, section("cases-0", "hero", heroData))

	// Checklist
	checklistData := map[string]interface{}{}
	if categories := cfg["categories"]; categories != nil {
		checklistData["categories"] = categories
	} else if cl := getMap(cfg, "checklist"); cl != nil {
		if cats := cl["categories"]; cats != nil {
			checklistData["categories"] = cats
		}
	}
	sections = append(sections, section("cases-1", "checklist", checklistData))

	return sections
}

func convertContact(cfg model.JSONMap) []map[string]interface{} {
	var sections []map[string]interface{}

	// Hero
	heroData := map[string]interface{}{}
	if hero := getMap(cfg, "hero"); hero != nil {
		copyFromMap(heroData, hero, "title", "subtitle", "backgroundColor")
	} else {
		copyIfPresent(heroData, cfg, "title", "subtitle", "backgroundColor")
	}
	sections = append(sections, section("contact-0", "hero", heroData))

	// Contact form
	formData := map[string]interface{}{}
	if form := getMap(cfg, "form"); form != nil {
		formData["form"] = form
	}
	if contact := getMap(cfg, "contact"); contact != nil {
		formData["contact"] = contact
	}
	if cf := getMap(cfg, "contactForm"); cf != nil {
		if form := cf["form"]; form != nil {
			formData["form"] = form
		}
		if contact := cf["contact"]; contact != nil {
			formData["contact"] = contact
		}
	}
	sections = append(sections, section("contact-1", "contact-form", formData))

	return sections
}

// copyIfPresent copies keys from a JSONMap into a target map if they exist.
func copyIfPresent(dst map[string]interface{}, src model.JSONMap, keys ...string) {
	for _, k := range keys {
		if v, ok := src[k]; ok {
			dst[k] = v
		}
	}
}

// copyFromMap copies keys from a generic map into a target map if they exist.
func copyFromMap(dst map[string]interface{}, src map[string]interface{}, keys ...string) {
	for _, k := range keys {
		if v, ok := src[k]; ok {
			dst[k] = v
		}
	}
}

// prettyJSON marshals a value to pretty-printed JSON (for debugging).
func prettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
