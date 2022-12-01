package internal

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type MPackSchemaDefinition struct {
	ProjectName  string            `yaml:"name"`
	Author       string            `yaml:"author"`
	Version      string            `yaml:"version"`
	Dependencies []MPackDependency `yaml:"dependencies"`
	SkipFiles    []string          `yaml:"skip"`
	Generators   GeneratorList     `yaml:"generators"`
	MPackPath    string            `yaml:"-"`
}

type GeneratorList []*Generator

type Generator struct {
	Name    string                  // The name of the plugin
	Path    *string                 // The path to the plugin
	Options *map[string]interface{} // The plugin's options
}

type MPackDependency struct {
	Source DependencySource
	Path   string
}

type DependencySource string

const (
	DependencySource_Git   DependencySource = "git"
	DependencySource_Local DependencySource = "local"
)

func (d *MPackDependency) UnmarshalYAML(value *yaml.Node) error {
	var input string
	if err := value.Decode(&input); err != nil {
		return err
	}

	tokens := strings.Split(input, ":")
	if tokens[0] == "core" {
		d.Path = "__core"
		d.Source = DependencySource_Local
		return nil
	}

	source := DependencySource(tokens[0])
	path := tokens[1]

	switch source {
	case DependencySource_Git, DependencySource_Local:
		break

	default:
		return fmt.Errorf("invalid dependency source: %s", source)
	}

	d.Path = path
	d.Source = source
	return nil
}

func (d *GeneratorList) UnmarshalYAML(value *yaml.Node) error {
	var input []map[string]interface{}
	if err := value.Decode(&input); err != nil {
		return err
	}

	for _, m := range input {
		for k, v := range m {
			values := v.(map[string]interface{})
			var path *string
			var opts *map[string]interface{}

			p, ok := values["path"].(string)
			if ok {
				path = &p
			}

			o, ok := values["options"].(map[string]interface{})
			if ok {
				opts = &o
			}

			*d = append(*d, &Generator{
				Name:    k,
				Path:    path,
				Options: opts,
			})
		}
	}

	return nil
}

func (m *MPackSchemaDefinition) ShouldSkip(path string) bool {
	for _, skip := range m.SkipFiles {
		if skip == path {
			return true
		}
	}

	return false
}
