package project

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/internal/nexema"
	"tomasweigenast.com/nexema/tool/internal/plugin"
)

// ProjectConfig represents the contents of a nexema.yaml file
type ProjectConfig struct {
	Version    int              `yaml:"version" json:"version"`                   // Builder version, required.
	Name       string           `yaml:"name,omitempty" json:"name,omitempty"`     // Name of the project
	Author     string           `yaml:"author,omitempty" json:"author,omitempty"` // The author of the project
	Skip       []string         `yaml:"skip,omitempty" json:"skip,omitempty"`     // Skipped files, as glob references
	Generators NexemaGenerators `yaml:"generators" json:"generators"`             // At least one generator
}

type NexemaGenerators map[string]NexemaGenerator
type NexemaGenerator struct {
	Options map[string]any `yaml:"options" json:"options"`
	BinPath string         `yaml:"bin,omitempty" json:"bin,omitempty"`
}

// GetPlugin returns the associated plugin.Plugin with the generator
func (self *NexemaGenerators) GetPlugin(name string) (*plugin.Plugin, error) {
	generator, ok := (*self)[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not defined in nexema.yaml", name)
	}

	// if binPath is not defined, try to search in .nexema file for well known plugins
	binPath := generator.BinPath
	var err error
	if len(binPath) == 0 {
		binPath, err = nexema.GetWellKnownPlugin(name)
		if err != nil {
			return nil, err
		}
	}
	return plugin.NewPlugin(name, binPath), nil
}
