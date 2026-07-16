package plugin

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"blotting-consultancy/internal/provider"
	pb "blotting-consultancy/pkg/pluginproto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockProviderClient implements pb.ProviderServiceClient for testing.
type mockProviderClient struct {
	initResp      *pb.InitResponse
	initErr       error
	shutdownErr   error
	savePath      string
	saveErr       string
	getData       []byte
	getError      string
	getErr        error
	deleteErr     string
	urlResult     string
	existsResult  bool
	existsErr     string
	searchResults []*pb.SearchResult
	searchTotal   int64
	searchErr     string
	suggestions   []string
	suggestErr    string
	indexErr      string
	removeErr     string
	rebuildErr    string
	notifyErr     string
	captchaErr    string
	httpResp      *pb.HTTPResponse
	httpErr       error
}

func (m *mockProviderClient) Initialize(_ context.Context, _ *pb.InitRequest, _ ...grpc.CallOption) (*pb.InitResponse, error) {
	if m.initErr != nil {
		return nil, m.initErr
	}
	if m.initResp != nil {
		return m.initResp, nil
	}
	return &pb.InitResponse{Success: true}, nil
}

func (m *mockProviderClient) Shutdown(_ context.Context, _ *pb.ShutdownRequest, _ ...grpc.CallOption) (*pb.ShutdownResponse, error) {
	return &pb.ShutdownResponse{}, m.shutdownErr
}

func (m *mockProviderClient) StorageSave(_ context.Context, _ *pb.StorageSaveRequest, _ ...grpc.CallOption) (*pb.StorageSaveResponse, error) {
	return &pb.StorageSaveResponse{Path: m.savePath, Error: m.saveErr}, nil
}

func (m *mockProviderClient) StorageGet(_ context.Context, _ *pb.StorageGetRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[pb.StorageChunk], error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &mockStorageGetStream{
		chunks: []*pb.StorageChunk{{Data: m.getData, Error: m.getError}},
	}, nil
}

type mockStorageGetStream struct {
	chunks []*pb.StorageChunk
	index  int
}

func (s *mockStorageGetStream) Recv() (*pb.StorageChunk, error) {
	if s.index >= len(s.chunks) {
		return nil, io.EOF
	}
	chunk := s.chunks[s.index]
	s.index++
	return chunk, nil
}

func (*mockStorageGetStream) Header() (metadata.MD, error) { return nil, nil }
func (*mockStorageGetStream) Trailer() metadata.MD         { return nil }
func (*mockStorageGetStream) CloseSend() error             { return nil }
func (*mockStorageGetStream) Context() context.Context     { return context.Background() }
func (*mockStorageGetStream) SendMsg(any) error            { return nil }
func (*mockStorageGetStream) RecvMsg(any) error            { return nil }

func (m *mockProviderClient) StorageDelete(_ context.Context, _ *pb.StorageDeleteRequest, _ ...grpc.CallOption) (*pb.StorageDeleteResponse, error) {
	return &pb.StorageDeleteResponse{Error: m.deleteErr}, nil
}

func (m *mockProviderClient) StorageURL(_ context.Context, _ *pb.StorageURLRequest, _ ...grpc.CallOption) (*pb.StorageURLResponse, error) {
	return &pb.StorageURLResponse{Url: m.urlResult}, nil
}

func (m *mockProviderClient) StorageExists(_ context.Context, _ *pb.StorageExistsRequest, _ ...grpc.CallOption) (*pb.StorageExistsResponse, error) {
	return &pb.StorageExistsResponse{Exists: m.existsResult, Error: m.existsErr}, nil
}

func (m *mockProviderClient) Search(_ context.Context, _ *pb.SearchRequest, _ ...grpc.CallOption) (*pb.SearchResponse, error) {
	return &pb.SearchResponse{Results: m.searchResults, Total: m.searchTotal, Error: m.searchErr}, nil
}

func (m *mockProviderClient) SearchSuggest(_ context.Context, _ *pb.SearchSuggestRequest, _ ...grpc.CallOption) (*pb.SearchSuggestResponse, error) {
	return &pb.SearchSuggestResponse{Suggestions: m.suggestions, Error: m.suggestErr}, nil
}

func (m *mockProviderClient) SearchIndex(_ context.Context, _ *pb.SearchIndexRequest, _ ...grpc.CallOption) (*pb.SearchIndexResponse, error) {
	return &pb.SearchIndexResponse{Error: m.indexErr}, nil
}

func (m *mockProviderClient) SearchRemove(_ context.Context, _ *pb.SearchRemoveRequest, _ ...grpc.CallOption) (*pb.SearchRemoveResponse, error) {
	return &pb.SearchRemoveResponse{Error: m.removeErr}, nil
}

func (m *mockProviderClient) SearchRebuild(_ context.Context, _ *pb.SearchRebuildRequest, _ ...grpc.CallOption) (*pb.SearchRebuildResponse, error) {
	return &pb.SearchRebuildResponse{Error: m.rebuildErr}, nil
}

func (m *mockProviderClient) Notify(_ context.Context, _ *pb.NotifyRequest, _ ...grpc.CallOption) (*pb.NotifyResponse, error) {
	return &pb.NotifyResponse{Error: m.notifyErr}, nil
}

func (m *mockProviderClient) CaptchaVerify(_ context.Context, _ *pb.CaptchaVerifyRequest, _ ...grpc.CallOption) (*pb.CaptchaVerifyResponse, error) {
	return &pb.CaptchaVerifyResponse{Error: m.captchaErr}, nil
}

func (m *mockProviderClient) HandleHTTP(_ context.Context, _ *pb.HTTPRequest, _ ...grpc.CallOption) (*pb.HTTPResponse, error) {
	if m.httpErr != nil {
		return nil, m.httpErr
	}
	if m.httpResp != nil {
		return m.httpResp, nil
	}
	return &pb.HTTPResponse{StatusCode: 200}, nil
}

// newTestHostWithMock creates a GRPCHost with a mock client for testing.
func newTestHostWithMock(mock *mockProviderClient) *GRPCHost {
	meta := &PluginMeta{
		ID:      "test-plugin",
		Name:    "Test Plugin",
		Version: "1.0.0",
	}
	host := NewGRPCHost(meta, "/fake/binary")
	host.rpcClient = mock
	return host
}

func TestStorageProxy_Save(t *testing.T) {
	mock := &mockProviderClient{savePath: "uploads/test.jpg"}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	path, err := proxy.Save(context.Background(), "test.jpg", bytes.NewReader([]byte("data")), 4)
	require.NoError(t, err)
	assert.Equal(t, "uploads/test.jpg", path)
}

func TestStorageProxy_Save_Error(t *testing.T) {
	mock := &mockProviderClient{saveErr: "upload failed"}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	_, err := proxy.Save(context.Background(), "test.jpg", bytes.NewReader([]byte("data")), 4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
}

func TestStorageProxy_Get(t *testing.T) {
	mock := &mockProviderClient{
		getData: []byte("hello world"),
	}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	reader, err := proxy.Get(context.Background(), "test.txt")
	require.NoError(t, err)

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)
	assert.Equal(t, "hello world", buf.String())
}

func TestStorageProxy_Delete(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	err := proxy.Delete(context.Background(), "test.txt")
	assert.NoError(t, err)
}

func TestStorageProxy_Delete_Error(t *testing.T) {
	mock := &mockProviderClient{deleteErr: "file not found"}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	err := proxy.Delete(context.Background(), "test.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestStorageProxy_URL(t *testing.T) {
	mock := &mockProviderClient{urlResult: "https://cdn.example.com/test.jpg"}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	url := proxy.URL("test.jpg")
	assert.Equal(t, "https://cdn.example.com/test.jpg", url)
}

func TestStorageProxy_Exists(t *testing.T) {
	mock := &mockProviderClient{existsResult: true}
	host := newTestHostWithMock(mock)

	proxy := host.AsStorageProvider()
	exists, err := proxy.Exists(context.Background(), "test.txt")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestSearchProxy_Search(t *testing.T) {
	mock := &mockProviderClient{
		searchResults: []*pb.SearchResult{
			{Id: 1, Type: "article", Title: "Test", Score: 0.95},
		},
		searchTotal: 1,
	}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	resp, err := proxy.Search(context.Background(), "test", "en", "article", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.Results, 1)
	assert.Equal(t, "Test", resp.Results[0].Title)
}

func TestSearchProxy_Suggest(t *testing.T) {
	mock := &mockProviderClient{suggestions: []string{"test1", "test2"}}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	suggestions, err := proxy.Suggest(context.Background(), "tes", "en", 5)
	require.NoError(t, err)
	assert.Equal(t, []string{"test1", "test2"}, suggestions)
}

func TestSearchProxy_IndexArticle(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	err := proxy.IndexArticle(context.Background(), 1, "en", "Title", "Body", "slug")
	assert.NoError(t, err)
}

func TestSearchProxy_IndexPage(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	err := proxy.IndexPage(context.Background(), 1, "en", "Title", "Body", "slug")
	assert.NoError(t, err)
}

func TestSearchProxy_RemoveFromIndex(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	err := proxy.RemoveFromIndex(context.Background(), "article", 1)
	assert.NoError(t, err)
}

func TestSearchProxy_RebuildIndex(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsSearchProvider()
	err := proxy.RebuildIndex(context.Background())
	assert.NoError(t, err)
}

func TestNotifierProxy_Notify(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsNotifierProvider()
	err := proxy.Notify(context.Background(), provider.NotifyEvent{
		Type:    "test",
		Subject: "Test",
		Body:    "Hello",
		Meta:    map[string]string{"key": "value"},
	})
	assert.NoError(t, err)
}

func TestNotifierProxy_Notify_Error(t *testing.T) {
	mock := &mockProviderClient{notifyErr: "send failed"}
	host := newTestHostWithMock(mock)

	proxy := host.AsNotifierProvider()
	err := proxy.Notify(context.Background(), provider.NotifyEvent{Type: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send failed")
}

func TestCaptchaProxy_Verify(t *testing.T) {
	mock := &mockProviderClient{}
	host := newTestHostWithMock(mock)

	proxy := host.AsCaptchaProvider()
	err := proxy.Verify(context.Background(), "token123", "1.2.3.4")
	assert.NoError(t, err)
}

func TestCaptchaProxy_Verify_Error(t *testing.T) {
	mock := &mockProviderClient{captchaErr: "invalid token"}
	host := newTestHostWithMock(mock)

	proxy := host.AsCaptchaProvider()
	err := proxy.Verify(context.Background(), "bad-token", "1.2.3.4")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestHandleHTTP(t *testing.T) {
	mock := &mockProviderClient{
		httpResp: &pb.HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"ok":true}`),
		},
	}
	host := newTestHostWithMock(mock)

	resp, err := host.HandleHTTP(context.Background(), &pb.HTTPRequest{
		Method: "GET",
		Path:   "/test",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(200), resp.StatusCode)
}

func TestHandleHTTP_PluginNotRunning(t *testing.T) {
	host := NewGRPCHost(&PluginMeta{ID: "test"}, "/fake")
	// rpcClient is nil (not started)

	_, err := host.HandleHTTP(context.Background(), &pb.HTTPRequest{Method: "GET"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestNewGRPCHost(t *testing.T) {
	meta := &PluginMeta{ID: "test", Name: "Test", Version: "1.0.0"}
	host := NewGRPCHost(meta, "/path/to/binary")

	assert.Equal(t, "test", host.meta.ID)
	assert.Equal(t, "/path/to/binary", host.binaryPath)
	assert.False(t, host.IsRunning())
}

func TestGRPCHost_StopWhenNotStarted(t *testing.T) {
	host := NewGRPCHost(&PluginMeta{ID: "test"}, "/fake")
	err := host.Stop()
	assert.NoError(t, err)
}

func TestProviderProxyConcurrentClientRemoval(t *testing.T) {
	host := newTestHostWithMock(&mockProviderClient{})
	proxy := host.AsNotifierProvider()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			_ = proxy.Notify(context.Background(), provider.NotifyEvent{Type: "race"})
		}
	}()

	for range 1000 {
		host.mu.Lock()
		client := host.rpcClient
		host.rpcClient = nil
		host.rpcClient = client
		host.mu.Unlock()
	}
	<-done
}

func TestGRPCHostDrainsActiveRPCsBeforeStop(t *testing.T) {
	host := newTestHostWithMock(&mockProviderClient{})
	_, release, err := host.acquireProviderClient()
	require.NoError(t, err)

	host.mu.Lock()
	host.stopping = true
	host.mu.Unlock()
	_, _, err = host.acquireProviderClient()
	require.Error(t, err)

	drained := make(chan struct{})
	go func() {
		host.waitForActiveRPCs(time.Second)
		close(drained)
	}()

	select {
	case <-drained:
		t.Fatal("active RPC drained before release")
	case <-time.After(20 * time.Millisecond):
	}
	release()
	select {
	case <-drained:
	case <-time.After(time.Second):
		t.Fatal("active RPC did not drain after release")
	}
}
