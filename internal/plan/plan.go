package plan

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
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
	Applyable       *bool  `json:"applyable"`
	Complete        *bool  `json:"complete"`
	Errored         bool   `json:"errored"`
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func Analyze(r io.Reader) (Report, error) {
	const maximumBytes = 64 << 20
	limited := &io.LimitedReader{R: r, N: maximumBytes + 1}
	dec := json.NewDecoder(limited)
	var doc document
	if err := dec.Decode(&doc); err != nil {
		if limited.N == 0 {
			return Report{}, fmt.Errorf("plan JSON exceeds %d bytes", maximumBytes)
		}
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
	major, err := formatMajor(doc.FormatVersion)
	if err != nil {
		return Report{}, err
	}
	if major != 1 {
		return Report{}, fmt.Errorf("unsupported format_version %q", doc.FormatVersion)
	}
	if limited.N == 0 {
		return Report{}, fmt.Errorf("plan JSON exceeds %d bytes", maximumBytes)
	}
	if doc.Applyable == nil || doc.Complete == nil {
		return Report{}, fmt.Errorf("plan JSON must include applyable and complete status")
	}
	if doc.Errored {
		return Report{}, fmt.Errorf("plan is marked errored")
	}
	if !*doc.Applyable {
		return Report{}, fmt.Errorf("plan is not applyable")
	}
	if !*doc.Complete {
		return Report{}, fmt.Errorf("plan is incomplete")
	}
	var report Report
	seen := make(map[string]struct{}, len(doc.ResourceChanges))
	for _, resource := range doc.ResourceChanges {
		if resource.Address == "" {
			return Report{}, fmt.Errorf("resource change has an empty address")
		}
		if _, exists := seen[resource.Address]; exists {
			return Report{}, fmt.Errorf("resource %s appears more than once", resource.Address)
		}
		seen[resource.Address] = struct{}{}
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
	sort.Strings(report.Destructive)
	return report, nil
}

func formatMajor(value string) (int, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, fmt.Errorf("invalid format_version %q", value)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid format_version %q", value)
	}
	if _, err := strconv.Atoi(parts[1]); err != nil {
		return 0, fmt.Errorf("invalid format_version %q", value)
	}
	return major, nil
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
