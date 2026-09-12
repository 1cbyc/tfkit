package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func CheckTree(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	dirs := []string{root}
	children, err := filepath.Glob(filepath.Join(root, "modules", "*"))
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if hasTerraformFiles(child) {
			dirs = append(dirs, child)
		}
	}
	var findings []string
	for _, dir := range dirs {
		moduleFindings, checkErr := checkModule(dir)
		if checkErr != nil {
			return nil, checkErr
		}
		findings = append(findings, moduleFindings...)
	}
	sort.Strings(findings)
	return findings, nil
}

func checkModule(dir string) ([]string, error) {
	var findings []string
	for _, required := range []string{"README.md", "main.tf", "variables.tf", "outputs.tf", "versions.tf"} {
		if _, err := os.Stat(filepath.Join(dir, required)); err != nil {
			if os.IsNotExist(err) {
				findings = append(findings, fmt.Sprintf("%s: missing %s", dir, required))
				continue
			}
			return nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		findings = append(findings, fmt.Sprintf("%s: no Terraform files", dir))
		return findings, nil
	}
	parser := hclparse.NewParser()
	var hasRequiredVersion bool
	for _, file := range files {
		parsed, diagnostics := parser.ParseHCLFile(file)
		if diagnostics.HasErrors() {
			return nil, fmt.Errorf("parse %s: %s", file, diagnostics.Error())
		}
		body, ok := parsed.Body.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("parse %s: unsupported body", file)
		}
		for _, block := range body.Blocks {
			switch block.Type {
			case "terraform":
				if attr, ok := block.Body.Attributes["required_version"]; ok {
					hasRequiredVersion = true
					if err := checkVersionConstraint(attr); err != nil {
						findings = append(findings, location(file, "required_version", err.Error()))
					}
				}
				for _, nested := range block.Body.Blocks {
					if nested.Type == "backend" {
						findings = append(findings, location(file, nested.Type, "reusable modules must not configure a backend"))
					}
				}
			case "variable", "output":
				if _, ok := block.Body.Attributes["description"]; !ok {
					findings = append(findings, location(file, block.Type, "missing description"))
				}
				if block.Type == "variable" {
					if _, ok := block.Body.Attributes["type"]; !ok {
						findings = append(findings, location(file, block.Type, "missing explicit type"))
					}
				}
			case "resource":
				for _, nested := range block.Body.Blocks {
					if nested.Type == "provisioner" {
						findings = append(findings, location(file, nested.Type, "provisioners are not allowed"))
					}
				}
			}
		}
	}
	if !hasRequiredVersion {
		findings = append(findings, fmt.Sprintf("%s: missing terraform.required_version", dir))
	}
	return findings, nil
}

func checkVersionConstraint(attr *hclsyntax.Attribute) error {
	value, diagnostics := attr.Expr.Value(nil)
	if diagnostics.HasErrors() || value.Type() != cty.String || !value.IsKnown() {
		return fmt.Errorf("must be a static string constraint")
	}
	constraints, err := version.NewConstraint(value.AsString())
	if err != nil {
		return fmt.Errorf("is invalid: %w", err)
	}
	minimum := version.Must(version.NewVersion("1.16.0"))
	nextMajor := version.Must(version.NewVersion("2.0.0"))
	if !constraints.Check(minimum) || constraints.Check(nextMajor) {
		return fmt.Errorf("must include Terraform 1.16.0 and exclude 2.0.0")
	}
	return nil
}

func hasTerraformFiles(dir string) bool {
	files, _ := filepath.Glob(filepath.Join(dir, "*.tf"))
	return len(files) > 0
}

func location(file, kind, message string) string {
	return fmt.Sprintf("%s: %s %s", file, kind, message)
}
