package module

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
)

// Module defines the contract for a self-contained feature module.
type Module interface {
	Name() string
	Init(deps Dependencies) error
	RegisterRoutes(public, admin *gin.RouterGroup)
}

// Dependencies provides shared resources that modules need.
type Dependencies struct {
	DB       *gorm.DB
	Registry *provider.Registry
	Repos    *SharedRepos
	SiteCfg  repository.SiteConfigRepository
}

// SharedRepos holds cross-module repositories.
type SharedRepos struct {
	ContentDoc repository.ContentDocumentRepository
	Article    repository.ArticleRepository
}
