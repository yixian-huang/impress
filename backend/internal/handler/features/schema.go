package features

import (
	"encoding/json"
	"errors"

	"blotting-consultancy/internal/model"
)

type PublicPages struct {
	Home         bool `json:"home"`
	Blog         bool `json:"blog"`
	Contact      bool `json:"contact"`
	About        bool `json:"about"`
	Experts      bool `json:"experts"`
	CoreServices bool `json:"coreServices"`
	Advantages   bool `json:"advantages"`
	Cases        bool `json:"cases"`
}

type BlogFeatures struct {
	Comments         bool `json:"comments"`
	RSS              bool `json:"rss"`
	ReadingMeta      bool `json:"readingMeta"`
	WordsPerMinute   int  `json:"wordsPerMinute,omitempty"`
}

type Features struct {
	PublicPages PublicPages  `json:"publicPages"`
	Blog        BlogFeatures `json:"blog"`
}

func validateFeatures(raw model.JSONMap) (*Features, error) {
	if raw == nil {
		return nil, errors.New("features payload required")
	}
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var f Features
	if err := json.Unmarshal(bytes, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
