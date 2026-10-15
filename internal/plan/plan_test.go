package plan

import (
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()
	report, err := Analyze(strings.NewReader(`{"format_version":"1.2","applyable":true,"complete":true,"resource_changes":[{"address":"a.one","change":{"actions":["create"]}},{"address":"b.two","change":{"actions":["update"]}},{"address":"d.four","change":{"actions":["delete"]}},{"address":"c.three","change":{"actions":["delete","create"]}},{"address":"e.five","change":{"actions":["no-op"]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if report.Create != 1 || report.Update != 1 || report.Replace != 1 || report.Delete != 1 || report.NoOp != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if got := strings.Join(report.Destructive, ","); got != "c.three,d.four" {
		t.Fatalf("unexpected destructive addresses: %s", got)
	}
}

func TestAnalyzeRejectsInvalidBoundaries(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"invalid JSON":       `{`,
		"missing version":    `{"applyable":true,"complete":true,"resource_changes":[]}`,
		"invalid version":    `{"format_version":"one","applyable":true,"complete":true}`,
		"new major version":  `{"format_version":"2.0","applyable":true,"complete":true}`,
		"missing status":     `{"format_version":"1.2","resource_changes":[]}`,
		"not applyable":      `{"format_version":"1.2","applyable":false,"complete":true}`,
		"incomplete":         `{"format_version":"1.2","applyable":true,"complete":false}`,
		"errored":            `{"format_version":"1.2","applyable":true,"complete":true,"errored":true}`,
		"trailing data":      `{"format_version":"1.2","applyable":true,"complete":true} true`,
		"unsupported action": `{"format_version":"1.2","applyable":true,"complete":true,"resource_changes":[{"address":"a.one","change":{"actions":["forget"]}}]}`,
		"empty address":      `{"format_version":"1.2","applyable":true,"complete":true,"resource_changes":[{"address":"","change":{"actions":["create"]}}]}`,
		"duplicate address":  `{"format_version":"1.2","applyable":true,"complete":true,"resource_changes":[{"address":"a.one","change":{"actions":["create"]}},{"address":"a.one","change":{"actions":["update"]}}]}`,
	}
	for name, input := range tests {
		input := input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Analyze(strings.NewReader(input)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAnalyzeAcceptsTerraformMetadata(t *testing.T) {
	t.Parallel()
	_, err := Analyze(strings.NewReader(`{"format_version":"1.99","terraform_version":"1.16.2","applyable":true,"complete":true,"variables":{},"future_property":{},"resource_changes":[]}`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeRejectsOversizedInput(t *testing.T) {
	t.Parallel()
	input := `{"format_version":"1.2","applyable":true,"complete":true,"padding":"` + strings.Repeat("x", (64<<20)+1) + `"}`
	if _, err := Analyze(strings.NewReader(input)); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size error, got %v", err)
	}
}
