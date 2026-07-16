package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	AIConfigSingletonID = uint(1)

	AIProviderOpenAI    = "openai"
	AIProviderAnthropic = "anthropic"
	AIProviderDisabled  = "disabled"
)

// AIConfig stores the singleton AI provider configuration.
type AIConfig struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Provider         string    `gorm:"not null;size:32;default:disabled" json:"provider"`
	APIKeyCiphertext string    `gorm:"type:text" json:"-"`
	BaseURL          string    `gorm:"size:500" json:"baseUrl,omitempty"`
	Model            string    `gorm:"size:255" json:"model,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (AIConfig) TableName() string {
	return "ai_configs"
}

func (c *AIConfig) Validate() error {
	switch c.Provider {
	case AIProviderOpenAI, AIProviderAnthropic, AIProviderDisabled:
		return nil
	default:
		return errors.New("provider must be one of: openai, anthropic, disabled")
	}
}

func (c *AIConfig) BeforeCreate(_ *gorm.DB) error {
	c.ID = AIConfigSingletonID
	return c.Validate()
}

func (c *AIConfig) BeforeSave(_ *gorm.DB) error {
	c.ID = AIConfigSingletonID
	return c.Validate()
}

func (c *AIConfig) HasAPIKey() bool {
	return c.APIKeyCiphertext != ""
}
