package nexema

import (
	"fmt"

	"tomasweigenast.com/nexema/tool/plugin"
)

type NexemaGenerators map[string]NexemaGenerator
type NexemaGenerator struct {
	Options map[string]any `yaml:"options" json:"options"`
	BinPath string         `yaml:"bin,omitempty" json:"bin,omitempty"`
}

func (self *NexemaGenerators) GetPlugin(name string) (*plugin.Plugin, error) {
	generator, ok := (*self)[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not defined in nexema.yaml", name)
	}

	// if binPath is not defined, try to search in .nexema file for well known plugins
	binPath := generator.BinPath
	var err error
	if len(binPath) == 0 {
		binPath, err = GetWellKnownPlugin(name)
		if err != nil {
			return nil, err
		}
	}
	return plugin.NewPlugin(name, binPath), nil
}
