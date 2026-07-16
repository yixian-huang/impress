package secretcipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// Prefix identifies the ciphertext format and key derivation version.
	Prefix = "v1:aes-gcm:"
)

var (
	ErrEmptySecret       = errors.New("server secret cannot be empty")
	ErrUnsupportedFormat = errors.New("unsupported ciphertext format")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Cipher encrypts and decrypts short application secrets.
type Cipher struct {
	key [32]byte
}

// New derives an AES-256 key from the supplied server secret using SHA-256.
func New(serverSecret string) (*Cipher, error) {
	serverSecret = strings.TrimSpace(serverSecret)
	if serverSecret == "" {
		return nil, ErrEmptySecret
	}

	return &Cipher{key: sha256.Sum256([]byte(serverSecret))}, nil
}

// Encrypt returns a version-prefixed AES-GCM ciphertext.
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm cipher: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return Prefix + base64.RawURLEncoding.EncodeToString(sealed), nil
}

// Decrypt opens a version-prefixed AES-GCM ciphertext.
func (c *Cipher) Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, Prefix) {
		return "", ErrUnsupportedFormat
	}

	encoded := strings.TrimPrefix(ciphertext, Prefix)
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCiphertext, err)
	}

	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return "", fmt.Errorf("create aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm cipher: %w", err)
	}
	if len(data) < gcm.NonceSize()+gcm.Overhead() {
		return "", ErrInvalidCiphertext
	}

	nonce := data[:gcm.NonceSize()]
	payload := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCiphertext, err)
	}
	return string(plaintext), nil
}
