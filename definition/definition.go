package definition

// NexemaFile represents a .nex file and its contents
type NexemaFile struct {
	Id          uint64           `json:"id"`          // The id of the file
	FileName    string           `json:"fileName"`    // The name of the file, without any path
	PackageName string           `json:"packageName"` // The name of the package
	Path        string           `json:"path"`        // The path to the file, relative to nexema.yaml
	Types       []TypeDefinition `json:"types"`       // The list of types defined
}

// NexemaSnapshot represents a generated project definition
type NexemaSnapshot struct {
	Version  int          `json:"version"` // The Nexema version, at the moment, always 1
	Hashcode uint64       `json:"hashcode"`
	Files    []NexemaFile `json:"files"`
}
