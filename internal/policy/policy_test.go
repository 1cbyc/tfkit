package policy

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGoodModulePasses(t *testing.T) {
	t.Parallel()
	findings, err := CheckTree(filepath.Join("..", "..", "tests", "fixtures", "module-good"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("unexpected findings: %v", findings)
	}
}

func TestBadModuleReportsFocusedFindings(t *testing.T) {
	t.Parallel()
	findings, err := CheckTree(filepath.Join("..", "..", "tests", "fixtures", "module-bad"))
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(findings, "\n")
	for _, want := range []string{"missing README.md", "missing outputs.tf", "missing explicit type", "provisioners are not allowed", "reusable modules must not configure a backend"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing finding %q in:\n%s", want, joined)
		}
	}
}
