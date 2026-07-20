package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

type memPVRepo struct {
	mu    sync.Mutex
	items []*model.PageView
}

func (m *memPVRepo) Create(ctx context.Context, pv *model.PageView) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *pv
	m.items = append(m.items, &cp)
	return nil
}

func (m *memPVRepo) CreateBatch(ctx context.Context, views []*model.PageView) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, pv := range views {
		cp := *pv
		m.items = append(m.items, &cp)
	}
	return nil
}

func (m *memPVRepo) GetSummary(ctx context.Context, now time.Time) ([]repository.PageViewStats, error) {
	return nil, nil
}
func (m *memPVRepo) CountByPageKey(ctx context.Context, pageKey string) (int64, error) {
	return 0, nil
}
func (m *memPVRepo) CountSince(ctx context.Context, since time.Time) (int64, error) {
	return 0, nil
}

func (m *memPVRepo) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.items)
}

func TestPageViewRecorderBatchesAndFlushes(t *testing.T) {
	repo := &memPVRepo{}
	rec := NewPageViewRecorder(repo)
	rec.flushEvery = 30 * time.Millisecond
	rec.batchSize = 10
	rec.Start()
	defer rec.Stop(time.Second)

	for i := 0; i < 3; i++ {
		rec.Track("home", "zh", "v1", "")
	}

	require.Eventually(t, func() bool {
		return repo.len() >= 3
	}, time.Second, 20*time.Millisecond)
}

func TestPageViewRecorderStopDrains(t *testing.T) {
	repo := &memPVRepo{}
	rec := NewPageViewRecorder(repo)
	rec.flushEvery = time.Hour // force drain via Stop
	rec.Start()

	rec.Track("about", "en", "v2", "https://x.test")
	rec.Stop(time.Second)

	require.GreaterOrEqual(t, repo.len(), 1)
}
