package comment

import (
	"github.com/gin-gonic/gin"

	"blotting-consultancy/internal/module"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
)

// Module is the self-contained comment feature module.
type Module struct {
	handler  *Handler
	antispam *AntiSpamService
}

// New creates a new comment module.
func New() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "comment" }

func (m *Module) Init(deps module.Dependencies) error {
	if err := deps.DB.AutoMigrate(&Comment{}); err != nil {
		return err
	}
	repo := newGormRepository(deps.DB)
	captcha := &provider.NoopCaptchaProvider{}
	m.antispam = newAntiSpamService(captcha)
	var contentDoc repository.ContentDocumentRepository
	if deps.Repos != nil {
		contentDoc = deps.Repos.ContentDoc
	}
	m.handler = &Handler{
		repo:           repo,
		antispam:       m.antispam,
		siteCfgRepo:    deps.SiteCfg,
		contentDocRepo: contentDoc,
	}
	return nil
}

func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.POST("/comments", m.handler.PublicCreate)
	public.GET("/comments", m.handler.PublicList)
	admin.GET("/comments", m.handler.AdminList)
	admin.PATCH("/comments/:id/status", m.handler.AdminUpdateStatus)
	admin.DELETE("/comments/:id", m.handler.AdminDelete)
	admin.PUT("/comments/:id/pin", m.handler.AdminPin)
	admin.POST("/comments/reply", m.handler.AdminReply)
}

// Stop shuts down background goroutines (antispam cleanup).
func (m *Module) Stop() {
	if m.antispam != nil {
		m.antispam.Stop()
	}
}
