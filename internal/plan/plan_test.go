package plan

import (
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()
	report, err := Analyze(strings.NewReader(`{"format_version":"1.2","resource_changes":[{"address":"a.one","change":{"actions":["create"]}},{"address":"b.two","change":{"actions":["update"]}},{"address":"c.three","change":{"actions":["delete","create"]}},{"address":"d.four","change":{"actions":["delete"]}},{"address":"e.five","change":{"actions":["no-op"]}}]}`))
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
		"missing version":    `{"resource_changes":[]}`,
		"trailing data":      `{"format_version":"1.2"} true`,
		"unsupported action": `{"format_version":"1.2","resource_changes":[{"address":"a.one","change":{"actions":["forget"]}}]}`,
		"empty address":      `{"format_version":"1.2","resource_changes":[{"address":"","change":{"actions":["create"]}}]}`,
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
	_, err := Analyze(strings.NewReader(`{"format_version":"1.2","terraform_version":"1.16.2","variables":{},"resource_changes":[]}`))
	if err != nil {
		t.Fatal(err)
	}
}
