package volumeleaders_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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
	if err != nil {
		t.Fatalf("external module import smoke test failed: %v\n%s", err, string(output))
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() = %v", err)
	}
	return filepath.Dir(cwd)
}

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) = %v", path, err)
	}
}
