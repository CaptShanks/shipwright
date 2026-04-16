package installer

import (
	"strings"
	"testing"
)

func TestMdAgentToTOML(t *testing.T) {
	md := []byte(`---
name: solution-architect
description: Designs solution architecture by analyzing the codebase
model: default
effort: high
skills:
  - security-awareness
  - scalability-resilience
---

## Role

You are a senior solution architect.
`)

	toml, err := mdAgentToTOML(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := string(toml)
	if !strings.Contains(got, `name = "solution-architect"`) {
		t.Errorf("expected name field, got:\n%s", got)
	}
	if !strings.Contains(got, `description = "Designs solution architecture by analyzing the codebase"`) {
		t.Errorf("expected description field, got:\n%s", got)
	}
	if !strings.Contains(got, `developer_instructions = `) {
		t.Errorf("expected developer_instructions field, got:\n%s", got)
	}
	if strings.Contains(got, "---") {
		t.Errorf("TOML should not contain frontmatter delimiters, got:\n%s", got)
	}
}

func TestMdAgentToTOML_NoFrontmatter(t *testing.T) {
	md := []byte("Just a body with no frontmatter")

	_, err := mdAgentToTOML(md)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestSplitFrontmatter(t *testing.T) {
	input := []byte("---\nname: test\n---\nBody here\n")
	fm, body, err := splitFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(fm, "name: test") {
		t.Errorf("expected frontmatter to contain 'name: test', got: %q", fm)
	}
	if !strings.Contains(body, "Body here") {
		t.Errorf("expected body to contain 'Body here', got: %q", body)
	}
}

func TestExtractYAMLField(t *testing.T) {
	yaml := "name: my-agent\ndescription: Does cool things\nmodel: default"

	if got := extractYAMLField(yaml, "name"); got != "my-agent" {
		t.Errorf("name: got %q, want %q", got, "my-agent")
	}
	if got := extractYAMLField(yaml, "description"); got != "Does cool things" {
		t.Errorf("description: got %q, want %q", got, "Does cool things")
	}
	if got := extractYAMLField(yaml, "missing"); got != "" {
		t.Errorf("missing: got %q, want empty", got)
	}
}
