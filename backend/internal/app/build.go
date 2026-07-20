package app

// BuildInfo carries ldflags-injected release metadata for /version and /health.
type BuildInfo struct {
	Version   string
	BuildTime string
	GitCommit string
	GitBranch string
}

// DefaultBuildInfo is used when the binary is not stamped with ldflags.
func DefaultBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   "dev",
		BuildTime: "unknown",
		GitCommit: "unknown",
		GitBranch: "unknown",
	}
}
