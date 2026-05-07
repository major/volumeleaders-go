package volumeleaders_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalImports(t *testing.T) {
	root := moduleRoot(t)
	dir := t.TempDir()
	goMod := "module volumeleaders_import_smoke\n\ngo 1.26.3\n\nrequire github.com/major/volumeleaders-go v0.0.0\n\nreplace github.com/major/volumeleaders-go => " + root + "\n"
	goTest := "package smoke\n\nimport (\n\t\"testing\"\n\n\tvolumeleaders \"github.com/major/volumeleaders-go/volumeleaders\"\n)\n\nfunc TestCoreExplicitSessionImport(t *testing.T) {\n\t_ = t\n\t_ = volumeleaders.NewClient\n\t_ = volumeleaders.NewSession\n\t_ = volumeleaders.SessionFromCookies\n}\n"
	mustWriteFile(t, filepath.Join(dir, "go.mod"), goMod)
	mustWriteFile(t, filepath.Join(dir, "smoke_test.go"), goTest)

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "external module import smoke test failed:\n%s", string(output))
}

func TestRootPackageDependencyIsolation(t *testing.T) {
	deps := listPackageDeps(t, "github.com/major/volumeleaders-go/volumeleaders")
	err := forbidDependencySubstrings(deps, []string{
		"github.com/browserutils",
		"kooky",
		"sqlite",
		"keyring",
		"dbus",
	})
	require.NoError(t, err)
}

func TestRootPackageDependencyIsolationFailsClosed(t *testing.T) {
	tests := []struct {
		name      string
		deps      []string
		wantError bool
	}{
		{
			name: "clean dependency list",
			deps: []string{
				"github.com/major/volumeleaders-go/volumeleaders",
				"net/http",
			},
			wantError: false,
		},
		{
			name: "forbidden dependency is rejected",
			deps: []string{
				"github.com/major/volumeleaders-go/volumeleaders",
				"github.com/browserutils/kooky",
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := forbidDependencySubstrings(tc.deps, []string{"github.com/browserutils", "kooky", "sqlite", "keyring", "dbus"})
			if tc.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "github.com/browserutils/kooky")
				return
			}

			require.NoError(t, err)
		})
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Dir(cwd)
}

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600), "WriteFile(%q)", path)
}

func listPackageDeps(t *testing.T, packagePath string) []string {
	t.Helper()

	root := moduleRoot(t)
	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", packagePath)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "go list -deps failed:\n%s", string(output))

	deps := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(deps) == 1 && deps[0] == "" {
		return nil
	}

	return deps
}

func forbidDependencySubstrings(deps []string, forbidden []string) error {
	for _, dep := range deps {
		for _, bad := range forbidden {
			if strings.Contains(dep, bad) {
				return fmt.Errorf("forbidden dependency %q found in %q", bad, dep)
			}
		}
	}

	return nil
}
