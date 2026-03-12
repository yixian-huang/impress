package model

import (
	"errors"
	"time"
)

// Glossary represents a bilingual glossary term used by the translation system
type Glossary struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SourceLang string    `gorm:"not null;size:10;index:idx_glossary_langs" json:"sourceLang"`
	TargetLang string    `gorm:"not null;size:10;index:idx_glossary_langs" json:"targetLang"`
	SourceTerm string    `gorm:"not null;size:500" json:"sourceTerm"`
	TargetTerm string    `gorm:"not null;size:500" json:"targetTerm"`
	Context    string    `gorm:"size:1000" json:"context"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// Validate validates the glossary model
func (g *Glossary) Validate() error {
	if g.SourceLang == "" {
		return errors.New("sourceLang is required")
	}
	if g.TargetLang == "" {
		return errors.New("targetLang is required")
	}
	if g.SourceTerm == "" {
		return errors.New("sourceTerm is required")
	}
	if g.TargetTerm == "" {
		return errors.New("targetTerm is required")
	}
	return nil
}
