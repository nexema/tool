package cmd

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// MPackSchemaConfig defines the parsing structure of the mpack.[yaml|json] file
type MPackSchemaConfig struct {
	Name         string                        `json:"name" yaml:"name"`
	Author       string                        `json:"author" yaml:"author"`
	Version      string                        `json:"version" yaml:"version"`
	Dependencies []MPackSchemaConfigDependency `json:"dependencies" yaml:"dependencies"`
	Skip         []string                      `json:"skip" yaml:"skip"`
	Generators   MPackSchemaConfigGeneratorMap `json:"generators" yaml:"generators"`
}

type MPackSchemaConfigDependency struct {
	Source MPackSchemaConfigDependencySource
	Path   string
}

type MPackSchemaConfigDependencySource int8

const (
	Git   MPackSchemaConfigDependencySource = 1
	Local MPackSchemaConfigDependencySource = 2
)

type MPackSchemaConfigGeneratorMap map[string]MPackSchemaConfigGenerator
type MPackSchemaConfigGenerator struct {
	Out     string   `json:"out" yaml:"out"`
	Options []string `json:"options" yaml:"options"`
}

func (m *MPackSchemaConfigDependency) UnmarshalJSON(b []byte) error {
	raw := string(b)
	raw = raw[1 : len(raw)-1] // this fix is really necessary?
	parts := strings.Split(raw, ":")

	if len(parts) != 2 {
		return fmt.Errorf("dependencies must be in the format source:path, given %s", raw)
	}

	source := strings.ToLower(parts[0])
	path := parts[1]

	if source == "git" {
		m.Source = Git
	} else if source == "local" {
		m.Source = Local
	} else {
		return fmt.Errorf("unknown dependency source %s", source)
	}

	m.Path = path
	return nil
}

func (m *MPackSchemaConfigDependency) UnmarshalYAML(b *yaml.Node) error {
	var raw string
	err := b.Decode(&raw)
	if err != nil {
		return err
	}

	parts := strings.Split(raw, ":")

	if len(parts) != 2 {
		return fmt.Errorf("dependencies must be in the format source:path, given %s", raw)
	}

	source := strings.ToLower(parts[0])
	path := parts[1]

	if source == "git" {
		m.Source = Git
	} else if source == "local" {
		m.Source = Local
	} else {
		return fmt.Errorf("unknown dependency source %s", source)
	}

	m.Path = path
	return nil
}

type FileDefinition struct {
	FileName string `json:"name"`
}
