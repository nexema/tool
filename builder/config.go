package builder

// NexemaConfig represents the contents of a nexema.yaml file
type NexemaConfig struct {
	Version    int               `yaml:"version" json:"version"`
	Name       string            `yaml:"name" json:"name"`
	Autor      string            `yaml:"author" json:"author"`
	Skip       []string          `yaml:"skip" json:"skip"` // skipped files, as glob references
	Generators []NexemaGenerator `yaml:"generators" json:"generators"`
}

type NexemaGenerator struct {
	Options []string `yaml:"options" json:"options"`
	BinPath string   `yaml:"bin" json:"bin"`
}
