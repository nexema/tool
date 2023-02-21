package definition

// NexemaFile represents a .nex file and its contents
type NexemaFile struct {
	FileName    string           `json:"fileName"`
	PackageName string           `json:"packageName"`
	Path        string           `json:"path"`
	Types       []TypeDefinition `json:"types"`
}

// NexemaDefinition represents a generated project definition
type NexemaDefinition struct {
	Version int          `json:"version"` // The Nexema version, at the moment, always 1
	Files   []NexemaFile `json:"files"`
}
