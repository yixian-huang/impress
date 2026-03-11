package seo

import (
	"encoding/json"
)

type BreadcrumbItem struct {
	Name string
	URL  string
}

func OrganizationJSONLD(name, url, logoURL string) string {
	data := map[string]interface{}{
		"@context": "https://schema.org",
		"@type":    "Organization",
		"name":     name,
		"url":      url,
		"logo":     logoURL,
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func ArticleJSONLD(title, description, url, image, datePublished, author string) string {
	data := map[string]interface{}{
		"@context":      "https://schema.org",
		"@type":         "Article",
		"headline":      title,
		"description":   description,
		"url":           url,
		"image":         image,
		"datePublished": datePublished,
		"author": map[string]interface{}{
			"@type": "Person",
			"name":  author,
		},
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func BreadcrumbJSONLD(items []BreadcrumbItem) string {
	listItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		listItems[i] = map[string]interface{}{
			"@type":    "ListItem",
			"position": i + 1,
			"name":     item.Name,
			"item":     item.URL,
		}
	}
	data := map[string]interface{}{
		"@context":        "https://schema.org",
		"@type":           "BreadcrumbList",
		"itemListElement": listItems,
	}
	b, _ := json.Marshal(data)
	return string(b)
}
