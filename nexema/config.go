package nexema

// NexemaProjectConfig represents the contents of a nexema.yaml file
type NexemaProjectConfig struct {
	Version    int              `yaml:"version" json:"version"` // Builder version, required.
	Name       string           `yaml:"name,omitempty" json:"name,omitempty"`
	Autor      string           `yaml:"author,omitempty" json:"author,omitempty"`
	Skip       []string         `yaml:"skip,omitempty" json:"skip,omitempty"` // skipped files, as glob references
	Generators NexemaGenerators `yaml:"generators" json:"generators"`         // At least one
}
