package definition

// NexemaFile represents a .nex file and its contents
type NexemaFile struct {
	Id          uint64           `json:"id"`
	FileName    string           `json:"fileName"`
	PackageName string           `json:"packageName"`
	Path        string           `json:"path"`
	Types       []TypeDefinition `json:"types"`
}

// NexemaSnapshot represents a generated project definition
type NexemaSnapshot struct {
	Version  int          `json:"version"` // The Nexema version, at the moment, always 1
	Hashcode uint64       `json:"hashcode"`
	Files    []NexemaFile `json:"files"`
}
