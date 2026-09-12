package plan

import (
	"encoding/json"
	"fmt"
	"io"
)

type Report struct {
	Create      int
	Update      int
	Delete      int
	Replace     int
	NoOp        int
	Destructive []string
}

type document struct {
	FormatVersion   string `json:"format_version"`
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func Analyze(r io.Reader) (Report, error) {
	dec := json.NewDecoder(io.LimitReader(r, 64<<20))
	var doc document
	if err := dec.Decode(&doc); err != nil {
		return Report{}, fmt.Errorf("decode plan JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Report{}, fmt.Errorf("decode plan JSON: multiple JSON values")
		}
		return Report{}, fmt.Errorf("decode plan JSON: trailing data: %w", err)
	}
	if doc.FormatVersion == "" {
		return Report{}, fmt.Errorf("missing format_version")
	}
	var report Report
	for _, resource := range doc.ResourceChanges {
		if resource.Address == "" {
			return Report{}, fmt.Errorf("resource change has an empty address")
		}
		switch actionsKey(resource.Change.Actions) {
		case "create":
			report.Create++
		case "update":
			report.Update++
		case "delete":
			report.Delete++
			report.Destructive = append(report.Destructive, resource.Address)
		case "delete,create", "create,delete":
			report.Replace++
			report.Destructive = append(report.Destructive, resource.Address)
		case "no-op", "read":
			report.NoOp++
		default:
			return Report{}, fmt.Errorf("resource %s has unsupported actions %q", resource.Address, resource.Change.Actions)
		}
	}
	return report, nil
}

func actionsKey(actions []string) string {
	if len(actions) == 1 {
		return actions[0]
	}
	if len(actions) == 2 {
		return actions[0] + "," + actions[1]
	}
	return ""
}
