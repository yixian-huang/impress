# Impress external plugin SDK

External plugin binaries are trusted server-side code. Manifest permissions
are capability declarations for validation and audit; they are not an OS
sandbox. The runtime is disabled by default and requires both a system
administrator and `ENABLE_EXTERNAL_PLUGINS=true`.

External Go plugins import:

```go
import (
    pb "blotting-consultancy/pkg/pluginproto"
    "blotting-consultancy/pkg/pluginsdk"
)
```

Implement `pluginproto.ProviderServiceServer` (embedding
`pluginproto.UnimplementedProviderServiceServer` is recommended), then call
`pluginsdk.Serve` from `main`.

Secret-bearing `settingsSchema` fields are rejected in the current beta until
encrypted persistence and response masking are available.

The current external runtime accepts canonical `notifier`, `search`, and
`captcha` provider registrations. External storage providers, dependencies,
custom routes, and frontend entries remain reserved until their ownership and
delivery paths are fully coordinated.

## Compatibility policy

The SDK and protocol are beta. Additive protobuf fields and RPCs keep protocol
version 1. Removing or changing an existing field/RPC, changing lifecycle
semantics, or changing the go-plugin handshake requires a new protocol version.
Deprecated fields remain readable for at least one protocol version.

The plugin package consumed by Impress is a zip containing:

- `plugin.yaml`
- an executable named after the manifest `id`, or named `plugin`

See `examples/plugins/file-notifier` for a complete notifier provider.
