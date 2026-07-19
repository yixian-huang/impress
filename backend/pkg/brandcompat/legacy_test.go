package brandcompat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvValuePrefersCanonicalVariable(t *testing.T) {
	t.Setenv(EnvFileVariable, "/new/.env")
	t.Setenv(LegacyEnvFileVariable, "/legacy/.env")

	assert.Equal(t, "/new/.env", EnvValue(EnvFileVariable, LegacyEnvFileVariable))
}

func TestEnvValueReadsLegacyVariableAsFallback(t *testing.T) {
	t.Setenv(EnvFileVariable, "")
	t.Setenv(LegacyEnvFileVariable, "/legacy/.env")

	assert.Equal(t, "/legacy/.env", EnvValue(EnvFileVariable, LegacyEnvFileVariable))
}
