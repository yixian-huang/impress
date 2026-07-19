// Package pluginsdk exposes the versioned process boundary used by
// external Inkless plugins.
package pluginsdk

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "github.com/yixian-huang/inkless/backend/pkg/pluginproto"
)

// Handshake is the canonical handshake shared by Inkless and plugin processes.
// Changing it is a breaking plugin protocol change.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "INKLESS_PLUGIN",
	MagicCookieValue: "inkless-cms-v1",
}

// ProviderPluginName is the go-plugin dispatch key for provider plugins.
const ProviderPluginName = "provider"

// GRPCProviderPlugin bridges hashicorp/go-plugin to the generated provider
// service. Impl is populated only in the external plugin process.
type GRPCProviderPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
	Impl pb.ProviderServiceServer
}

func (p *GRPCProviderPlugin) GRPCServer(_ *goplugin.GRPCBroker, server *grpc.Server) error {
	pb.RegisterProviderServiceServer(server, p.Impl)
	return nil
}

func (p *GRPCProviderPlugin) GRPCClient(
	_ context.Context,
	_ *goplugin.GRPCBroker,
	connection *grpc.ClientConn,
) (interface{}, error) {
	return pb.NewProviderServiceClient(connection), nil
}

// Serve starts an external Inkless plugin process.
func Serve(implementation pb.ProviderServiceServer) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]goplugin.Plugin{
			ProviderPluginName: &GRPCProviderPlugin{Impl: implementation},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
