package service

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/draw"
)

// ProcessingJobStatus represents the status of a media processing job
type ProcessingJobStatus string

const (
	JobStatusPending    ProcessingJobStatus = "pending"
	JobStatusProcessing ProcessingJobStatus = "processing"
	JobStatusCompleted  ProcessingJobStatus = "completed"
	JobStatusFailed     ProcessingJobStatus = "failed"
)

// SizeVariant defines an image size variant to generate
type SizeVariant struct {
	Name      string
	MaxWidth  int
	MaxHeight int
}

// DefaultSizeVariants are the standard size variants generated
var DefaultSizeVariants = []SizeVariant{
	{Name: "thumb", MaxWidth: 150, MaxHeight: 150},
	{Name: "small", MaxWidth: 400, MaxHeight: 400},
	{Name: "medium", MaxWidth: 800, MaxHeight: 800},
	{Name: "large", MaxWidth: 1200, MaxHeight: 1200},
}

// ProcessingJob represents a single image processing job
type ProcessingJob struct {
	ID         string              `json:"id"`
	SourcePath string              `json:"sourcePath"`
	OutputDir  string              `json:"outputDir"`
	Status     ProcessingJobStatus `json:"status"`
	Error      string              `json:"error,omitempty"`
	Variants   []string            `json:"variants,omitempty"`
	CreatedAt  time.Time           `json:"createdAt"`
	UpdatedAt  time.Time           `json:"updatedAt"`
}

// MediaPipeline processes images asynchronously with bounded concurrency
type MediaPipeline struct {
	queue       chan *ProcessingJob
	jobs        map[string]*ProcessingJob
	mu          sync.RWMutex
	workerCount int
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewMediaPipeline creates a new media processing pipeline
func NewMediaPipeline(workerCount int) *MediaPipeline {
	if workerCount <= 0 {
		workerCount = 2
	}
	return &MediaPipeline{
		queue:       make(chan *ProcessingJob, 10), // bounded channel with capacity 10
		jobs:        make(map[string]*ProcessingJob),
		workerCount: workerCount,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the worker goroutines
func (p *MediaPipeline) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop gracefully shuts down the pipeline
func (p *MediaPipeline) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// Submit adds a new processing job to the queue. Returns false if the queue is full.
func (p *MediaPipeline) Submit(job *ProcessingJob) bool {
	job.Status = JobStatusPending
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	p.mu.Lock()
	p.jobs[job.ID] = job
	p.mu.Unlock()

	select {
	case p.queue <- job:
		return true
	default:
		// Queue is full - apply backpressure
		p.mu.Lock()
		job.Status = JobStatusFailed
		job.Error = "processing queue is full, try again later"
		job.UpdatedAt = time.Now()
		p.mu.Unlock()
		return false
	}
}

// GetJobStatus returns the current status of a processing job
func (p *MediaPipeline) GetJobStatus(id string) (*ProcessingJob, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	job, ok := p.jobs[id]
	if !ok {
		return nil, false
	}
	// Return a copy
	copy := *job
	return &copy, true
}

// worker processes jobs from the queue
func (p *MediaPipeline) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stopCh:
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			p.processJob(job)
		}
	}
}

// processJob processes a single image processing job
func (p *MediaPipeline) processJob(job *ProcessingJob) {
	p.mu.Lock()
	job.Status = JobStatusProcessing
	job.UpdatedAt = time.Now()
	p.mu.Unlock()

	variants, err := p.generateVariants(job.SourcePath, job.OutputDir)

	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err.Error()
	} else {
		job.Status = JobStatusCompleted
		job.Variants = variants
	}
	job.UpdatedAt = time.Now()
}

// generateVariants creates resized versions and a WebP-like optimized version
func (p *MediaPipeline) generateVariants(sourcePath, outputDir string) ([]string, error) {
	// Open source image
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	var variants []string

	for _, sv := range DefaultSizeVariants {
		outPath, err := p.resizeAndSave(srcImg, outputDir, baseName, sv)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s variant: %w", sv.Name, err)
		}
		variants = append(variants, outPath)
	}

	return variants, nil
}

// resizeAndSave resizes an image and saves it
func (p *MediaPipeline) resizeAndSave(src image.Image, outputDir, baseName string, sv SizeVariant) (string, error) {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	newW, newH := fitDimensions(srcW, srcH, sv.MaxWidth, sv.MaxHeight)

	// Skip if source is already smaller
	if srcW <= sv.MaxWidth && srcH <= sv.MaxHeight {
		newW = srcW
		newH = srcH
	}

	// Create resized image
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Save as JPEG (good compression, widely compatible)
	outPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.jpg", baseName, sv.Name))
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}

	return outPath, nil
}

// SaveAsPNG saves an image as PNG (used for transparency-preserving variants)
func SaveAsPNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// fitDimensions calculates dimensions that fit within maxW x maxH while maintaining aspect ratio
func fitDimensions(srcW, srcH, maxW, maxH int) (int, int) {
	if srcW <= 0 || srcH <= 0 {
		return maxW, maxH
	}

	ratioW := float64(maxW) / float64(srcW)
	ratioH := float64(maxH) / float64(srcH)

	ratio := ratioW
	if ratioH < ratioW {
		ratio = ratioH
	}

	newW := int(float64(srcW) * ratio)
	newH := int(float64(srcH) * ratio)

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	return newW, newH
}
