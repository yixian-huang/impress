package pluginsdk_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSDKBuildsFromIndependentModule(t *testing.T) {
	if testing.Short() {
		t.Skip("builds an independent Go module")
	}

	moduleRoot := repositoryModuleRoot(t)
	externalDir := t.TempDir()
	goMod := "module external-plugin\n\ngo 1.25.0\n\nrequire github.com/yixian-huang/inkless/backend v0.0.0\n\nreplace github.com/yixian-huang/inkless/backend => " + moduleRoot + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(externalDir, "go.mod"), []byte(goMod), 0o640))
	require.NoError(t, os.WriteFile(filepath.Join(externalDir, "main.go"), []byte(`package main

import (
	pb "github.com/yixian-huang/inkless/backend/pkg/pluginproto"
	"github.com/yixian-huang/inkless/backend/pkg/pluginsdk"
)

type server struct {
	pb.UnimplementedProviderServiceServer
}

func main() {
	pluginsdk.Serve(&server{})
}
`), 0o640))

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = externalDir
	output, err := tidy.CombinedOutput()
	require.NoError(t, err, string(output))

	build := exec.Command("go", "build", "./...")
	build.Dir = externalDir
	output, err = build.CombinedOutput()
	require.NoError(t, err, string(output))
}

func repositoryModuleRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
