package nexema

type WellKnownPlugin struct {
	Version      string   `json:"version"`
	PackageName  string   `json:"packageName"`
	InstallSteps []string `json:"steps"`
	BinaryName   string   `json:"binary"`
}

type PluginInfo struct {
	Name    string // the name of the plugin
	Version string // the installed version
	Path    string // the path to the binary
}
