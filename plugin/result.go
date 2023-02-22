package plugin

// PluginResult is the output of a Nexema generator plugin
type PluginResult struct {
	Success bool            `json:"success"`
	Files   []GeneratedFile `json:"files"`
}

type GeneratedFile struct {
	Id       uint64 `json:"id"`       // The id of the generated file
	Contents string `json:"contents"` // The contents of the file
}
