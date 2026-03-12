package model

import (
	"time"
)

// ChunkedUploadStatus represents the state of a chunked upload
type ChunkedUploadStatus string

const (
	ChunkedUploadPending    ChunkedUploadStatus = "pending"
	ChunkedUploadUploading  ChunkedUploadStatus = "uploading"
	ChunkedUploadCompleted  ChunkedUploadStatus = "completed"
	ChunkedUploadFailed     ChunkedUploadStatus = "failed"
)

// ChunkedUpload tracks the state of a chunked file upload
type ChunkedUpload struct {
	ID           string              `gorm:"primaryKey;size:64" json:"id"`
	Filename     string              `gorm:"not null;size:255" json:"filename"`
	MimeType     string              `gorm:"not null;size:100" json:"mimeType"`
	TotalSize    int64               `gorm:"not null" json:"totalSize"`
	TotalChunks  int                 `gorm:"not null" json:"totalChunks"`
	ChunkSize    int64               `gorm:"not null" json:"chunkSize"`
	UploadedChunks int               `gorm:"not null;default:0" json:"uploadedChunks"`
	Status       ChunkedUploadStatus `gorm:"not null;size:20;default:pending" json:"status"`
	TempDir      string              `gorm:"not null;size:500" json:"-"`
	CreatedAt    time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	ExpiresAt    time.Time           `gorm:"not null" json:"expiresAt"`
}

// TableName overrides the default table name
func (ChunkedUpload) TableName() string {
	return "chunked_uploads"
}
