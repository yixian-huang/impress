package seo

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sync"
	"time"
)

// Renderer renders index.html with dynamic meta tags.
// When created from a file path, Render() reloads the template if the file mtime changes
// (so frontend artifact deploys that share FRONTEND_DIR don't keep stale asset hashes).
type Renderer struct {
	mu       sync.Mutex
	tmpl     *template.Template
	filePath string
	mtime    time.Time
}

// NewRenderer creates a renderer from an index.html file path.
func NewRenderer(filePath string) (*Renderer, error) {
	r := &Renderer{filePath: filePath}
	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

// NewRendererFromString creates a renderer from a template string (no disk reload).
func NewRendererFromString(tmplStr string) (*Renderer, error) {
	t, err := template.New("index").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	return &Renderer{tmpl: t}, nil
}

func (r *Renderer) reload() error {
	content, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}
	info, err := os.Stat(r.filePath)
	if err != nil {
		return fmt.Errorf("stat template: %w", err)
	}
	t, err := template.New("index").Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	r.tmpl = t
	r.mtime = info.ModTime()
	return nil
}

func (r *Renderer) ensureFresh() {
	if r.filePath == "" {
		return
	}
	info, err := os.Stat(r.filePath)
	if err != nil {
		return
	}
	if info.ModTime().After(r.mtime) {
		_ = r.reload()
	}
}

// Render executes the template with the given PageMeta.
func (r *Renderer) Render(meta PageMeta) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensureFresh()
	if r.tmpl == nil {
		return "", fmt.Errorf("seo template not loaded")
	}
	var buf bytes.Buffer
	if err := r.tmpl.Execute(&buf, meta); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
