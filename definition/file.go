package definition

// NexemaFile represents a .nex file and its contents
type NexemaFile struct {
	FileName    string
	PackageName string
	Path        string
	Types       []TypeDefinition
}
