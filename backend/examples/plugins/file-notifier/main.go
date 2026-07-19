package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "github.com/yixian-huang/inkless/backend/pkg/pluginproto"
	"github.com/yixian-huang/inkless/backend/pkg/pluginsdk"
)

type fileNotifier struct {
	pb.UnimplementedProviderServiceServer

	mu         sync.RWMutex
	outputPath string
}

type notification struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	Meta      map[string]string `json:"meta,omitempty"`
}

func (n *fileNotifier) Initialize(_ context.Context, request *pb.InitRequest) (*pb.InitResponse, error) {
	outputPath := request.Settings["outputFile"]
	if outputPath == "" {
		outputPath = "notifications.jsonl"
	}
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(request.DataDir, outputPath)
	}

	cleanDataDir, err := filepath.Abs(request.DataDir)
	if err != nil {
		return &pb.InitResponse{Error: err.Error()}, nil
	}
	cleanOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return &pb.InitResponse{Error: err.Error()}, nil
	}
	relative, err := filepath.Rel(cleanDataDir, cleanOutputPath)
	if err != nil || relative == ".." || len(relative) > 3 && relative[:3] == ".."+string(filepath.Separator) {
		return &pb.InitResponse{Error: "outputFile must stay inside the plugin data directory"}, nil
	}
	if err := os.MkdirAll(filepath.Dir(cleanOutputPath), 0o750); err != nil {
		return &pb.InitResponse{Error: err.Error()}, nil
	}

	n.mu.Lock()
	n.outputPath = cleanOutputPath
	n.mu.Unlock()
	return &pb.InitResponse{Success: true}, nil
}

func (n *fileNotifier) Shutdown(context.Context, *pb.ShutdownRequest) (*pb.ShutdownResponse, error) {
	return &pb.ShutdownResponse{}, nil
}

func (n *fileNotifier) Notify(_ context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	n.mu.RLock()
	outputPath := n.outputPath
	n.mu.RUnlock()
	if outputPath == "" {
		return &pb.NotifyResponse{Error: "plugin is not initialized"}, nil
	}

	data, err := json.Marshal(notification{
		Timestamp: time.Now().UTC(),
		Type:      request.Type,
		Subject:   request.Subject,
		Body:      request.Body,
		Meta:      request.Meta,
	})
	if err != nil {
		return &pb.NotifyResponse{Error: err.Error()}, nil
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return &pb.NotifyResponse{Error: err.Error()}, nil
	}
	if _, err := fmt.Fprintf(file, "%s\n", data); err != nil {
		_ = file.Close()
		return &pb.NotifyResponse{Error: err.Error()}, nil
	}
	if err := file.Close(); err != nil {
		return &pb.NotifyResponse{Error: err.Error()}, nil
	}
	return &pb.NotifyResponse{}, nil
}

func (n *fileNotifier) HandleHTTP(context.Context, *pb.HTTPRequest) (*pb.HTTPResponse, error) {
	return &pb.HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"status":"ready"}`),
	}, nil
}

func main() {
	pluginsdk.Serve(&fileNotifier{})
}
