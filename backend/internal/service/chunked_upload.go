package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

// ChunkedUploadService manages chunked file uploads
type ChunkedUploadService struct {
	repo           repository.ChunkedUploadRepository
	mediaRepo      repository.MediaRepository
	tempDir        string
	uploadDir      string
	baseURL        string
	storageRuntime *StorageRuntimeService
}

// NewChunkedUploadService creates a new chunked upload service
func NewChunkedUploadService(
	repo repository.ChunkedUploadRepository,
	mediaRepo repository.MediaRepository,
	tempDir string,
	uploadDir string,
	baseURL string,
) *ChunkedUploadService {
	return NewChunkedUploadServiceWithStorage(
		repo,
		mediaRepo,
		tempDir,
		uploadDir,
		baseURL,
		ConfigureDefaultStorageRuntime(nil, uploadDir),
	)
}

func NewChunkedUploadServiceWithStorage(
	repo repository.ChunkedUploadRepository,
	mediaRepo repository.MediaRepository,
	tempDir string,
	uploadDir string,
	baseURL string,
	storageRuntime *StorageRuntimeService,
) *ChunkedUploadService {
	if storageRuntime == nil {
		storageRuntime = ConfigureDefaultStorageRuntime(nil, uploadDir)
	}
	return &ChunkedUploadService{
		repo:           repo,
		mediaRepo:      mediaRepo,
		tempDir:        tempDir,
		uploadDir:      uploadDir,
		baseURL:        baseURL,
		storageRuntime: storageRuntime,
	}
}

// InitUploadRequest holds the parameters for initializing a chunked upload
type InitUploadRequest struct {
	Filename    string `json:"filename"`
	MimeType    string `json:"mimeType"`
	TotalSize   int64  `json:"totalSize"`
	TotalChunks int    `json:"totalChunks"`
	ChunkSize   int64  `json:"chunkSize"`
}

// InitUpload initializes a new chunked upload session
func (s *ChunkedUploadService) InitUpload(ctx context.Context, req *InitUploadRequest) (*model.ChunkedUpload, error) {
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if req.TotalSize <= 0 {
		return nil, fmt.Errorf("totalSize must be positive")
	}
	if req.TotalChunks <= 0 {
		return nil, fmt.Errorf("totalChunks must be positive")
	}
	if req.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be positive")
	}

	uploadID := generateUploadID()
	uploadTempDir := filepath.Join(s.tempDir, uploadID)
	if err := os.MkdirAll(uploadTempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	upload := &model.ChunkedUpload{
		ID:          uploadID,
		Filename:    req.Filename,
		MimeType:    req.MimeType,
		TotalSize:   req.TotalSize,
		TotalChunks: req.TotalChunks,
		ChunkSize:   req.ChunkSize,
		Status:      model.ChunkedUploadPending,
		TempDir:     uploadTempDir,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.Create(ctx, upload); err != nil {
		os.RemoveAll(uploadTempDir)
		return nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	return upload, nil
}

// UploadChunk saves a single chunk
func (s *ChunkedUploadService) UploadChunk(ctx context.Context, uploadID string, chunkIndex int, data io.Reader) error {
	upload, err := s.repo.FindByID(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("upload not found: %w", err)
	}

	if upload.Status == model.ChunkedUploadCompleted {
		return fmt.Errorf("upload already completed")
	}

	if chunkIndex < 0 || chunkIndex >= upload.TotalChunks {
		return fmt.Errorf("invalid chunk index: %d (total chunks: %d)", chunkIndex, upload.TotalChunks)
	}

	chunkPath := filepath.Join(upload.TempDir, fmt.Sprintf("chunk_%06d", chunkIndex))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("failed to create chunk file: %w", err)
	}
	defer chunkFile.Close()

	if _, err := io.Copy(chunkFile, data); err != nil {
		os.Remove(chunkPath)
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	upload.UploadedChunks++
	upload.Status = model.ChunkedUploadUploading
	if err := s.repo.Update(ctx, upload); err != nil {
		return fmt.Errorf("failed to update upload record: %w", err)
	}

	return nil
}

// CompleteUpload merges all chunks and creates a media record
func (s *ChunkedUploadService) CompleteUpload(ctx context.Context, uploadID string) (*model.Media, error) {
	upload, err := s.repo.FindByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	if upload.Status == model.ChunkedUploadCompleted {
		return nil, fmt.Errorf("upload already completed")
	}

	// Read and sort chunk files
	entries, err := os.ReadDir(upload.TempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp directory: %w", err)
	}

	var chunkFiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			chunkFiles = append(chunkFiles, entry.Name())
		}
	}
	sort.Strings(chunkFiles)

	if len(chunkFiles) != upload.TotalChunks {
		return nil, fmt.Errorf("expected %d chunks, found %d", upload.TotalChunks, len(chunkFiles))
	}

	// Generate unique filename
	ext := filepath.Ext(upload.Filename)
	uniqueName := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), sanitizeFilenameForChunked(upload.Filename), ext)
	destPath := filepath.Join(upload.TempDir, uniqueName)

	// Merge chunks
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	var totalWritten int64
	for _, chunkName := range chunkFiles {
		chunkPath := filepath.Join(upload.TempDir, chunkName)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			os.Remove(destPath)
			return nil, fmt.Errorf("failed to open chunk %s: %w", chunkName, err)
		}
		n, err := io.Copy(destFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			os.Remove(destPath)
			return nil, fmt.Errorf("failed to merge chunk %s: %w", chunkName, err)
		}
		totalWritten += n
	}

	if _, err := destFile.Seek(0, io.SeekStart); err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to rewind merged file: %w", err)
	}
	storageKey, storageProvider, url, err := s.storageRuntime.Save(ctx, uniqueName, destFile, totalWritten)
	if err != nil {
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to save merged file: %w", err)
	}
	destFile.Close()

	media := &model.Media{
		URL:             url,
		Filename:        upload.Filename,
		MimeType:        upload.MimeType,
		Size:            totalWritten,
		StorageKey:      storageKey,
		StorageProvider: storageProvider,
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		_ = s.storageRuntime.Delete(ctx, storageProvider, storageKey)
		os.Remove(destPath)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	// Update upload status
	upload.Status = model.ChunkedUploadCompleted
	_ = s.repo.Update(ctx, upload)

	// Clean up temp directory
	os.RemoveAll(upload.TempDir)

	return media, nil
}

func generateUploadID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sanitizeFilenameForChunked(name string) string {
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	nameOnly := base[:len(base)-len(ext)]

	var result []byte
	for _, r := range nameOnly {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, byte(r))
		}
	}
	s := string(result)
	if len(s) > 50 {
		s = s[:50]
	}
	if s == "" {
		s = "file"
	}
	return s
}
