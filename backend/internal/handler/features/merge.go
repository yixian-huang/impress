package features

import "blotting-consultancy/internal/model"

// MergePublishedDefaults fills siteMode and ensures blog/publicPages shapes for API consumers.
func MergePublishedDefaults(raw model.JSONMap) model.JSONMap {
	if raw == nil {
		return nil
	}
	out := make(model.JSONMap)
	for k, v := range raw {
		out[k] = v
	}

	switch out["siteMode"] {
	case "blog", "corporate":
	default:
		if inferBlogSiteMode(out) {
			out["siteMode"] = "blog"
		} else {
			out["siteMode"] = "corporate"
		}
	}

	return out
}

func inferBlogSiteMode(raw model.JSONMap) bool {
	pp, ok := raw["publicPages"].(map[string]interface{})
	if !ok {
		return false
	}
	boolVal := func(k string) bool {
		v, ok := pp[k].(bool)
		return ok && v
	}
	return boolVal("blog") && boolVal("home") &&
		!boolVal("about") && !boolVal("experts") && !boolVal("coreServices") &&
		!boolVal("advantages") && !boolVal("cases")
}
