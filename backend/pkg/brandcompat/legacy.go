// Package brandcompat contains narrowly scoped compatibility identifiers used
// to read pre-Inkless configuration. Canonical identifiers always take priority.
package brandcompat

import (
	"log"
	"os"
	"strings"
	"sync"

	goplugin "github.com/hashicorp/go-plugin"
)

const (
	EnvFileVariable       = "INKLESS_ENV_FILE"
	SecretKeyVariable     = "INKLESS_SECRET_KEY"
	LegacyEnvFileVariable = "IMPRESS_ENV_FILE"
	LegacySecretVariable  = "IMPRESS_SECRET_KEY"
	LegacyStorageProbeKey = ".impress-storage-probe"
	LegacyWebhookEvent    = "X-Impress-Event"
	LegacyWebhookTime     = "X-Impress-Timestamp"
	LegacyWebhookSig      = "X-Impress-Signature"
)

// LegacyPluginHandshake contains the deprecated Impress plugin handshake.
// Deprecated: retained only so the plugin host can load existing plugins
// during the Inkless migration window. New plugins must use pluginsdk.Handshake.
var LegacyPluginHandshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "IMPRESS_PLUGIN",
	MagicCookieValue: "impress-cms-v1",
}

var warnedLegacyVariables sync.Map

// EnvValue reads a canonical environment variable and falls back to its
// deprecated predecessor. It never writes the deprecated variable.
func EnvValue(canonical, legacy string) string {
	if value := strings.TrimSpace(os.Getenv(canonical)); value != "" {
		return value
	}
	value := strings.TrimSpace(os.Getenv(legacy))
	if value == "" {
		return ""
	}
	if _, loaded := warnedLegacyVariables.LoadOrStore(legacy, struct{}{}); !loaded {
		log.Printf("deprecated environment variable %s is set; use %s instead", legacy, canonical)
	}
	return value
}
