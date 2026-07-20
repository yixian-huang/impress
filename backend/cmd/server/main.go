package main

import (
	"fmt"
	"os"

	_ "github.com/yixian-huang/inkless/backend/docs/swagger" // swagger docs
	"github.com/yixian-huang/inkless/backend/internal/app"
	"github.com/yixian-huang/inkless/backend/pkg/config"
)

// Build-time variables (set via ldflags).
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

// @title           Inkless CMS API
// @version         1.0
// @description     Extensible Inkless CMS API for content, themes, media, plugins, and site management.
// @host            localhost:8088
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" for JWT authentication
func main() {
	loadResult, err := config.LoadWithBootstrap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	application, err := app.New(loadResult, app.Options{
		Build: app.BuildInfo{
			Version:   Version,
			BuildTime: BuildTime,
			GitCommit: GitCommit,
			GitBranch: GitBranch,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
