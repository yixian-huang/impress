package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	goplugin "github.com/hashicorp/go-plugin"

	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/pkg/brandcompat"
	pb "github.com/yixian-huang/inkless/backend/pkg/pluginproto"
	"github.com/yixian-huang/inkless/backend/pkg/pluginsdk"
)

// GRPCHost manages a single plugin process via hashicorp/go-plugin.
type GRPCHost struct {
	meta       *PluginMeta
	binaryPath string
	client     *goplugin.Client
	rpcClient  pb.ProviderServiceClient
	mu         sync.RWMutex
	activeRPCs sync.WaitGroup
	stopping   bool
}

const (
	pluginStartTimeout    = 15 * time.Second
	pluginRPCTimeout      = 15 * time.Second
	pluginShutdownTimeout = 5 * time.Second
)

// NewGRPCHost creates a host for a plugin binary.
func NewGRPCHost(meta *PluginMeta, binaryPath string) *GRPCHost {
	return &GRPCHost{
		meta:       meta,
		binaryPath: binaryPath,
	}
}

// Start launches the plugin process and establishes gRPC connection.
func (h *GRPCHost) Start(settings map[string]string, dataDir string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.client != nil {
		return fmt.Errorf("plugin %s is already running", h.meta.ID)
	}
	handshake := pluginsdk.Handshake
	if usesLegacy, err := binaryUsesLegacyPluginHandshake(h.binaryPath); err != nil {
		log.Printf("plugin %s handshake marker preflight failed, using Inkless handshake: %v", h.meta.ID, err)
	} else if usesLegacy {
		handshake = brandcompat.LegacyPluginHandshake
		log.Printf("plugin %s uses legacy plugin handshake marker (compatibility path)", h.meta.ID)
	}

	client, rpcClient, err := h.startClientWithHandshake(handshake)
	if err != nil {
		killPluginClient(client)
		return fmt.Errorf("failed to create plugin client for %s: %w", h.meta.ID, err)
	}

	raw, err := rpcClient.Dispense(pluginsdk.ProviderPluginName)
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin %s: %w", h.meta.ID, err)
	}

	svc, ok := raw.(pb.ProviderServiceClient)
	if !ok {
		client.Kill()
		return fmt.Errorf("plugin %s did not return ProviderServiceClient", h.meta.ID)
	}

	// Initialize the plugin with settings
	ctx, cancel := context.WithTimeout(context.Background(), pluginRPCTimeout)
	defer cancel()
	resp, err := svc.Initialize(ctx, &pb.InitRequest{
		Settings: settings,
		DataDir:  dataDir,
		PluginId: h.meta.ID,
	})
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to initialize plugin %s: %w", h.meta.ID, err)
	}
	if !resp.Success {
		client.Kill()
		return fmt.Errorf("plugin %s initialization failed: %s", h.meta.ID, resp.Error)
	}

	h.client = client
	h.rpcClient = svc
	h.stopping = false
	return nil
}

func (h *GRPCHost) startClientWithHandshake(handshake goplugin.HandshakeConfig) (*goplugin.Client, goplugin.ClientProtocol, error) {
	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]goplugin.Plugin{
			pluginsdk.ProviderPluginName: &pluginsdk.GRPCProviderPlugin{},
		},
		Cmd:              exec.Command(h.binaryPath),
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		StartTimeout:     pluginStartTimeout,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, err
	}

	return client, rpcClient, nil
}

func killPluginClient(client *goplugin.Client) {
	if client != nil {
		client.Kill()
	}
}

func binaryUsesLegacyPluginHandshake(binaryPath string) (bool, error) {
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return false, err
	}
	hasCanonical := bytes.Contains(data, []byte(pluginsdk.Handshake.MagicCookieKey)) &&
		bytes.Contains(data, []byte(pluginsdk.Handshake.MagicCookieValue))
	hasLegacy := bytes.Contains(data, []byte(brandcompat.LegacyPluginHandshake.MagicCookieKey)) &&
		bytes.Contains(data, []byte(brandcompat.LegacyPluginHandshake.MagicCookieValue))
	return hasLegacy && !hasCanonical, nil
}

// Stop gracefully shuts down the plugin process.
func (h *GRPCHost) Stop() error {
	h.mu.Lock()
	if h.client == nil {
		h.mu.Unlock()
		return nil
	}
	if h.stopping {
		h.mu.Unlock()
		return nil
	}
	h.stopping = true
	client := h.client
	svc := h.rpcClient
	h.mu.Unlock()

	h.waitForActiveRPCs(pluginShutdownTimeout)
	if svc != nil {
		ctx, cancel := context.WithTimeout(context.Background(), pluginShutdownTimeout)
		_, _ = svc.Shutdown(ctx, &pb.ShutdownRequest{})
		cancel()
	}

	client.Kill()
	h.mu.Lock()
	h.client = nil
	h.rpcClient = nil
	h.stopping = false
	h.mu.Unlock()
	return nil
}

// IsRunning checks if the plugin process is alive.
func (h *GRPCHost) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client != nil && !h.client.Exited()
}

func (h *GRPCHost) acquireProviderClient() (pb.ProviderServiceClient, func(), error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rpcClient == nil || h.stopping {
		return nil, nil, fmt.Errorf("plugin %s is not running", h.meta.ID)
	}
	h.activeRPCs.Add(1)
	return h.rpcClient, h.activeRPCs.Done, nil
}

func (h *GRPCHost) waitForActiveRPCs(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		h.activeRPCs.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

// Reinitialize applies settings to a running plugin without persisting them.
func (h *GRPCHost) Reinitialize(ctx context.Context, settings map[string]string, dataDir string) error {
	svc, release, err := h.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	callCtx, cancel := context.WithTimeout(ctx, pluginRPCTimeout)
	defer cancel()
	resp, err := svc.Initialize(callCtx, &pb.InitRequest{
		Settings: settings,
		DataDir:  dataDir,
		PluginId: h.meta.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to re-initialize plugin %s: %w", h.meta.ID, err)
	}
	if !resp.Success {
		return fmt.Errorf("plugin %s re-initialization failed: %s", h.meta.ID, resp.Error)
	}
	return nil
}

// AsStorageProvider returns a StorageProvider that proxies to the plugin.
func (h *GRPCHost) AsStorageProvider() provider.StorageProvider {
	return &grpcStorageProxy{host: h}
}

// AsSearchProvider returns a SearchProvider that proxies to the plugin.
func (h *GRPCHost) AsSearchProvider() provider.SearchProvider {
	return &grpcSearchProxy{host: h}
}

// AsNotifierProvider returns a NotifierProvider that proxies to the plugin.
func (h *GRPCHost) AsNotifierProvider() provider.NotifierProvider {
	return &grpcNotifierProxy{host: h}
}

// AsCaptchaProvider returns a CaptchaProvider that proxies to the plugin.
func (h *GRPCHost) AsCaptchaProvider() provider.CaptchaProvider {
	return &grpcCaptchaProxy{host: h}
}

// HandleHTTP proxies an HTTP request to the plugin.
func (h *GRPCHost) HandleHTTP(ctx context.Context, req *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	svc, release, err := h.acquireProviderClient()
	if err != nil {
		return nil, err
	}
	defer release()
	return svc.HandleHTTP(ctx, req)
}

// --- Provider proxy types ---

// grpcStorageProxy implements provider.StorageProvider by proxying to a plugin via gRPC.
type grpcStorageProxy struct {
	host *GRPCHost
}

func (p *grpcStorageProxy) Save(ctx context.Context, filename string, reader io.Reader, size int64) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return "", err
	}
	defer release()
	resp, err := svc.StorageSave(ctx, &pb.StorageSaveRequest{
		Filename: filename,
		Data:     data,
		Size:     size,
	})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.Path, nil
}

func (p *grpcStorageProxy) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return nil, err
	}
	defer release()
	stream, err := svc.StorageGet(ctx, &pb.StorageGetRequest{Path: path})
	if err != nil {
		return nil, err
	}
	var data bytes.Buffer
	for {
		chunk, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if chunk.Error != "" {
			return nil, fmt.Errorf("%s", chunk.Error)
		}
		if _, err := data.Write(chunk.Data); err != nil {
			return nil, err
		}
	}
	return io.NopCloser(bytes.NewReader(data.Bytes())), nil
}

func (p *grpcStorageProxy) Delete(ctx context.Context, path string) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.StorageDelete(ctx, &pb.StorageDeleteRequest{Path: path})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (p *grpcStorageProxy) URL(path string) string {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return ""
	}
	defer release()
	ctx, cancel := context.WithTimeout(context.Background(), pluginRPCTimeout)
	defer cancel()
	resp, err := svc.StorageURL(ctx, &pb.StorageURLRequest{Path: path})
	if err != nil {
		return ""
	}
	return resp.Url
}

func (p *grpcStorageProxy) Exists(ctx context.Context, path string) (bool, error) {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return false, err
	}
	defer release()
	resp, err := svc.StorageExists(ctx, &pb.StorageExistsRequest{Path: path})
	if err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, fmt.Errorf("%s", resp.Error)
	}
	return resp.Exists, nil
}

// grpcSearchProxy implements provider.SearchProvider by proxying to a plugin via gRPC.
type grpcSearchProxy struct {
	host *GRPCHost
}

func (p *grpcSearchProxy) Search(ctx context.Context, query string, locale string, contentType string, page int, pageSize int) (*provider.SearchResponse, error) {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return nil, err
	}
	defer release()
	resp, err := svc.Search(ctx, &pb.SearchRequest{
		Query:       query,
		Locale:      locale,
		ContentType: contentType,
		Page:        int32(page),
		PageSize:    int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	results := make([]provider.SearchResult, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = provider.SearchResult{
			ID:      uint(r.Id),
			Type:    r.Type,
			Title:   r.Title,
			Snippet: r.Snippet,
			URL:     r.Url,
			Locale:  r.Locale,
			Score:   r.Score,
		}
	}
	return &provider.SearchResponse{
		Results:  results,
		Total:    resp.Total,
		Page:     page,
		PageSize: pageSize,
		Query:    query,
	}, nil
}

func (p *grpcSearchProxy) Suggest(ctx context.Context, prefix string, locale string, limit int) ([]string, error) {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return nil, err
	}
	defer release()
	resp, err := svc.SearchSuggest(ctx, &pb.SearchSuggestRequest{
		Prefix: prefix,
		Locale: locale,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	return resp.Suggestions, nil
}

func (p *grpcSearchProxy) IndexArticle(ctx context.Context, id uint, locale string, title string, body string, slug string) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.SearchIndex(ctx, &pb.SearchIndexRequest{
		ContentType: "article",
		Id:          uint64(id),
		Locale:      locale,
		Title:       title,
		Body:        body,
		Slug:        slug,
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (p *grpcSearchProxy) IndexPage(ctx context.Context, id uint, locale string, title string, body string, slug string) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.SearchIndex(ctx, &pb.SearchIndexRequest{
		ContentType: "page",
		Id:          uint64(id),
		Locale:      locale,
		Title:       title,
		Body:        body,
		Slug:        slug,
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (p *grpcSearchProxy) RemoveFromIndex(ctx context.Context, contentType string, id uint) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.SearchRemove(ctx, &pb.SearchRemoveRequest{
		ContentType: contentType,
		Id:          uint64(id),
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (p *grpcSearchProxy) RebuildIndex(ctx context.Context) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.SearchRebuild(ctx, &pb.SearchRebuildRequest{})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// grpcNotifierProxy implements provider.NotifierProvider by proxying to a plugin via gRPC.
type grpcNotifierProxy struct {
	host *GRPCHost
}

func (p *grpcNotifierProxy) Notify(ctx context.Context, event provider.NotifyEvent) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.Notify(ctx, &pb.NotifyRequest{
		Type:    event.Type,
		Subject: event.Subject,
		Body:    event.Body,
		Meta:    event.Meta,
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// grpcCaptchaProxy implements provider.CaptchaProvider by proxying to a plugin via gRPC.
type grpcCaptchaProxy struct {
	host *GRPCHost
}

func (p *grpcCaptchaProxy) Verify(ctx context.Context, token string, remoteIP string) error {
	svc, release, err := p.host.acquireProviderClient()
	if err != nil {
		return err
	}
	defer release()
	resp, err := svc.CaptchaVerify(ctx, &pb.CaptchaVerifyRequest{
		Token:    token,
		RemoteIp: remoteIP,
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
