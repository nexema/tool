package nexema

type WellKnownPlugin struct {
	Version      string   `json:"version"`
	PackageName  string   `json:"packageName"`
	InstallSteps []string `json:"steps"`
	BinaryName   string   `json:"binary"`
}

type PluginInfo struct {
	Name    string
	Version string
	Path    string
}
