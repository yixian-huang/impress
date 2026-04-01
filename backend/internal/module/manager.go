package module

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Manager manages the lifecycle of feature modules.
type Manager struct {
	modules []Module
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Register(mod Module) {
	m.modules = append(m.modules, mod)
}

func (m *Manager) InitAll(deps Dependencies) error {
	for _, mod := range m.modules {
		if err := mod.Init(deps); err != nil {
			return fmt.Errorf("module %s init: %w", mod.Name(), err)
		}
	}
	return nil
}

func (m *Manager) RegisterAllRoutes(public, admin *gin.RouterGroup) {
	for _, mod := range m.modules {
		mod.RegisterRoutes(public, admin)
	}
}

func (m *Manager) Has(name string) bool {
	for _, mod := range m.modules {
		if mod.Name() == name {
			return true
		}
	}
	return false
}
