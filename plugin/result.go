package plugin

// PluginResult is the output of a Nexema generator plugin
type PluginResult struct {
	ExitCode int             `json:"exitCode"`
	Files    []GeneratedFile `json:"files"`
}

type GeneratedFile struct {
	Id       uint64 `json:"id"`       // The id of the generated file
	Name     string `json:"name"`     // The name of the file
	Contents string `json:"contents"` // The contents of the file
}
