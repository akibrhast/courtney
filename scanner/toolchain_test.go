package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akibrhast/courtney/patsy/vos"
	"github.com/akibrhast/courtney/shared"
)

// Exercise the installed Go toolchain's standard-library export data. Older
// x/tools versions can build successfully but fail to load newer export data.
func TestLoadStandardLibraryTypes(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	files := map[string]string{
		"go.mod": "module example.com/toolchaincheck\n\ngo 1.25.0\n",
		"check.go": `package toolchaincheck
import "context"
func Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

`,
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cm := New(&shared.Setup{
		Env:      vos.Os(),
		Packages: []shared.PackageSpec{{Path: "example.com/toolchaincheck", Dir: dir}},
	})
	if err := cm.LoadProgram(); err != nil {
		t.Fatal(err)
	}
	if len(cm.pkgs) != 1 {
		t.Fatalf("expected one loaded package, got %d", len(cm.pkgs))
	}
	pkg := cm.pkgs[0]
	if len(pkg.Errors) != 0 {
		t.Fatalf("package loading errors: %v", pkg.Errors)
	}
	ctx := pkg.Imports["context"]
	if ctx == nil || ctx.Types == nil || !ctx.Types.Complete() {
		t.Fatal("expected complete standard-library context types")
	}
}

func TestLoadProgramReportsTypeErrors(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	for name, contents := range map[string]string{
		"go.mod":    "module example.com/broken\n\ngo 1.26.0\n",
		"broken.go": "package broken\nvar Value = undefinedName\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cm := New(&shared.Setup{
		Env:      vos.Os(),
		Packages: []shared.PackageSpec{{Path: "example.com/broken", Dir: dir}},
	})
	err := cm.LoadProgram()
	if err == nil || !strings.Contains(err.Error(), "undefinedName") {
		t.Fatalf("expected the package type error, got %v", err)
	}
}
