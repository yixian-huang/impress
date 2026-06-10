package setup

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/seed"
	"blotting-consultancy/pkg/auth"
)

var (
	ErrAlreadyCompleted = errors.New("setup already completed")
	ErrInvalidInput     = errors.New("invalid setup input")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// AdminInput is the administrator account for first-time setup.
type AdminInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SiteInput is basic site metadata collected during setup.
type SiteInput struct {
	Name          map[string]string `json:"name"`
	DefaultLocale string            `json:"defaultLocale"`
}

// CompleteInput is the payload for POST /setup/complete.
type CompleteInput struct {
	Admin    AdminInput `json:"admin"`
	Site     SiteInput  `json:"site"`
	SeedMode string     `json:"seedMode"`
}

// Status describes whether the instance has completed web setup.
type Status struct {
	Installed    bool   `json:"installed"`
	DatabaseType string `json:"databaseType"`
}

// Service handles first-run installation via the web wizard.
type Service struct {
	userRepo    repository.UserRepository
	siteCfgRepo repository.SiteConfigRepository
	contentRepo repository.ContentDocumentRepository
	seeder      *seed.Seeder
	seedRBAC    func(ctx context.Context) error
}

// NewService creates a setup service.
func NewService(
	userRepo repository.UserRepository,
	siteCfgRepo repository.SiteConfigRepository,
	contentRepo repository.ContentDocumentRepository,
	seeder *seed.Seeder,
	seedRBAC func(ctx context.Context) error,
) *Service {
	return &Service{
		userRepo:    userRepo,
		siteCfgRepo: siteCfgRepo,
		contentRepo: contentRepo,
		seeder:      seeder,
		seedRBAC:    seedRBAC,
	}
}

// IsInstalled reports whether setup has already been completed.
func (s *Service) IsInstalled(ctx context.Context) (bool, error) {
	if installed, err := s.hasSystemInstallRecord(ctx); err != nil {
		return false, err
	} else if installed {
		return true, nil
	}

	count, err := s.userRepo.CountSuperAdmins(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetStatus returns install state plus database type hint for the wizard UI.
func (s *Service) GetStatus(ctx context.Context, databaseType string) (*Status, error) {
	installed, err := s.IsInstalled(ctx)
	if err != nil {
		return nil, err
	}
	return &Status{
		Installed:    installed,
		DatabaseType: databaseType,
	}, nil
}

// Complete runs the one-shot installation flow.
func (s *Service) Complete(ctx context.Context, in CompleteInput) error {
	installed, err := s.IsInstalled(ctx)
	if err != nil {
		return err
	}
	if installed {
		return ErrAlreadyCompleted
	}
	if err := validateInput(in); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	existing, _ := s.userRepo.FindByUsername(ctx, strings.TrimSpace(in.Admin.Username))
	if existing != nil {
		return fmt.Errorf("%w: username already exists", ErrInvalidInput)
	}

	hashedPassword, err := auth.HashPassword(in.Admin.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Username:     strings.TrimSpace(in.Admin.Username),
		PasswordHash: hashedPassword,
		Role:         model.RoleAdmin,
		IsSuperAdmin: true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(in.SeedMode)) {
	case "demo":
		if err := s.seeder.DemoSiteSeedContent(ctx); err != nil {
			return fmt.Errorf("demo seed: %w", err)
		}
	default:
		if err := s.seeder.BlankSiteSeedContent(ctx); err != nil {
			return fmt.Errorf("blank seed: %w", err)
		}
	}

	if err := applyGlobalSiteName(ctx, s.contentRepo, in.Site); err != nil {
		return fmt.Errorf("apply site name: %w", err)
	}

	seedMode := strings.ToLower(strings.TrimSpace(in.SeedMode))
	if seedMode != "demo" {
		seedMode = "blank"
	}
	installCfg := model.JSONMap{
		"installed":        true,
		"installedAt":      time.Now().UTC().Format(time.RFC3339),
		"seedMode":         seedMode,
		"installerVersion": "1",
	}
	row := &model.SiteConfig{
		Key:              model.SiteConfigKeySystem,
		DraftConfig:      installCfg,
		DraftVersion:     1,
		PublishedConfig:  installCfg,
		PublishedVersion: 1,
	}
	if err := s.siteCfgRepo.Upsert(ctx, row); err != nil {
		return fmt.Errorf("write install record: %w", err)
	}

	if s.seedRBAC != nil {
		if err := s.seedRBAC(ctx); err != nil {
			return fmt.Errorf("seed rbac: %w", err)
		}
	}

	return nil
}

func (s *Service) hasSystemInstallRecord(ctx context.Context) (bool, error) {
	sc, err := s.siteCfgRepo.FindByKey(ctx, model.SiteConfigKeySystem)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "record not found") {
			return false, nil
		}
		return false, err
	}
	if sc == nil || sc.ID == 0 {
		return false, nil
	}
	installed, _ := sc.PublishedConfig["installed"].(bool)
	return installed, nil
}

func validateInput(in CompleteInput) error {
	username := strings.TrimSpace(in.Admin.Username)
	if !usernamePattern.MatchString(username) {
		return errors.New("username must be 3-32 characters (letters, digits, underscore)")
	}
	if err := validatePassword(in.Admin.Password); err != nil {
		return err
	}

	nameZH := strings.TrimSpace(in.Site.Name["zh"])
	nameEN := strings.TrimSpace(in.Site.Name["en"])
	if nameZH == "" && nameEN == "" {
		return errors.New("site name is required")
	}

	locale := strings.TrimSpace(in.Site.DefaultLocale)
	if locale != "" && locale != "zh" && locale != "en" {
		return errors.New("defaultLocale must be zh or en")
	}

	mode := strings.ToLower(strings.TrimSpace(in.SeedMode))
	if mode != "" && mode != "blank" && mode != "demo" {
		return errors.New("seedMode must be blank or demo")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("password must contain letters and digits")
	}
	return nil
}

func applyGlobalSiteName(ctx context.Context, contentRepo repository.ContentDocumentRepository, site SiteInput) error {
	doc, err := contentRepo.FindByPageKey(ctx, model.PageKeyGlobal)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "record not found") {
			return nil
		}
		return err
	}
	if doc == nil {
		return nil
	}

	cfg := doc.PublishedConfig
	if cfg == nil {
		cfg = model.JSONMap{}
	}

	identity, _ := cfg["identity"].(model.JSONMap)
	if identity == nil {
		identity = model.JSONMap{}
	}

	nameMap := model.JSONMap{}
	if v := strings.TrimSpace(site.Name["zh"]); v != "" {
		nameMap["zh"] = v
	}
	if v := strings.TrimSpace(site.Name["en"]); v != "" {
		nameMap["en"] = v
	}
	if len(nameMap) > 0 {
		identity["name"] = nameMap
	}

	locale := strings.TrimSpace(site.DefaultLocale)
	if locale == "" {
		locale = "zh"
	}
	identity["defaultLocale"] = locale
	switch locale {
	case "en":
		identity["localeMode"] = "mono-en"
	case "zh":
		identity["localeMode"] = "mono-zh"
	default:
		identity["localeMode"] = "bilingual"
	}

	cfg["identity"] = identity
	doc.DraftConfig = cfg
	doc.PublishedConfig = cfg
	if doc.DraftVersion < 1 {
		doc.DraftVersion = 1
	}
	if doc.PublishedVersion < 1 {
		doc.PublishedVersion = 1
	}

	return contentRepo.Update(ctx, doc)
}
