package project

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
