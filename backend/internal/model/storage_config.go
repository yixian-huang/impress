package model

import (
	"errors"
	"time"
)

// StorageStrategy represents the type of storage backend
type StorageStrategy string

const (
	StorageLocal StorageStrategy = "local"
	StorageS3    StorageStrategy = "s3"
	StorageOSS   StorageStrategy = "oss"
)

// StorageConfig holds the storage configuration
type StorageConfig struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Strategy  StorageStrategy `gorm:"not null;size:20;default:local" json:"strategy"`
	Bucket    string          `gorm:"size:255" json:"bucket,omitempty"`
	Region    string          `gorm:"size:100" json:"region,omitempty"`
	Endpoint  string          `gorm:"size:500" json:"endpoint,omitempty"`
	AccessKey string          `gorm:"size:255" json:"accessKey,omitempty"`
	SecretKey string          `gorm:"size:255" json:"-"`
	BasePath  string          `gorm:"size:500" json:"basePath,omitempty"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName overrides the default table name
func (StorageConfig) TableName() string {
	return "storage_configs"
}

// Validate validates the storage config
func (s *StorageConfig) Validate() error {
	switch s.Strategy {
	case StorageLocal:
		// No additional fields required for local storage
	case StorageS3, StorageOSS:
		if s.Bucket == "" {
			return errors.New("bucket is required for s3/oss storage")
		}
		if s.AccessKey == "" {
			return errors.New("accessKey is required for s3/oss storage")
		}
		if s.SecretKey == "" {
			return errors.New("secretKey is required for s3/oss storage")
		}
		if s.Strategy == StorageOSS && s.Endpoint == "" {
			return errors.New("endpoint is required for oss storage")
		}
	default:
		return errors.New("strategy must be one of: local, s3, oss")
	}
	return nil
}

// HasSecretKey returns true if a secret key is set (for API responses)
func (s *StorageConfig) HasSecretKey() bool {
	return s.SecretKey != ""
}
