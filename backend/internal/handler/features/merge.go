package features

import "github.com/yixian-huang/inkless/backend/internal/model"

// MergePublishedDefaults normalizes published features for bootstrap.
// Legacy siteMode is stripped; active theme is the single source of truth for site presentation.
func MergePublishedDefaults(raw model.JSONMap) model.JSONMap {
	if raw == nil {
		return nil
	}
	out := make(model.JSONMap)
	for k, v := range raw {
		if k == "siteMode" {
			continue
		}
		out[k] = v
	}
	return out
}
