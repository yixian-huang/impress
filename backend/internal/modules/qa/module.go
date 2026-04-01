package qa

import (
	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/module"
)

type Module struct {
	handler *Handler
}

func New() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "qa" }

func (m *Module) Init(deps module.Dependencies) error {
	if err := deps.DB.AutoMigrate(&QALog{}); err != nil {
		return err
	}
	qaLogRepo := newGormQALogRepository(deps.DB)
	vectorStore := NewMemoryVectorStore()
	qaService := NewQAService(deps.Registry.AI(), vectorStore)
	embeddingService := NewEmbeddingService(deps.Registry.AI(), vectorStore)
	m.handler = &Handler{
		qaService:        qaService,
		embeddingService: embeddingService,
		qaLogRepo:        qaLogRepo,
		contentDocRepo:   deps.Repos.ContentDoc,
		articleRepo:      deps.Repos.Article,
		siteCfgRepo:      deps.SiteCfg,
	}
	return nil
}

func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.POST("/qa/ask", m.handler.PublicAsk)
	admin.POST("/qa/index", m.handler.AdminIndex)
	admin.GET("/qa/logs", m.handler.AdminListLogs)
	admin.POST("/qa/logs/:id/feedback", m.handler.AdminFeedback)
}
