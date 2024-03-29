package plugin

// PluginResult is the output of a Nexema generator plugin
type PluginResult struct {
	ExitCode int              `json:"exitCode"`
	Files    *[]GeneratedFile `json:"files"`
	Error    *string          `json:"error"`
}

type GeneratedFile struct {
	Id       string `json:"id"`       // The id of the generated file
	Name     string `json:"name"`     // The name of the file
	Contents string `json:"contents"` // The contents of the file
}
