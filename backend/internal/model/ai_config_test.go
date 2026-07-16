package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAIConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{name: "openai", provider: AIProviderOpenAI},
		{name: "anthropic", provider: AIProviderAnthropic},
		{name: "disabled", provider: AIProviderDisabled},
		{name: "invalid", provider: "other", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&AIConfig{Provider: tt.provider}).Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAIConfig_BeforeSaveForcesSingletonID(t *testing.T) {
	config := &AIConfig{ID: 99, Provider: AIProviderOpenAI}
	require.NoError(t, config.BeforeSave(nil))
	require.Equal(t, AIConfigSingletonID, config.ID)
}
