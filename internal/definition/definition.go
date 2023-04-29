package definition

// NexemaFile represents a .nex file and its contents
type NexemaFile struct {
	Id          string           `json:"id"`          // The id of the file
	PackageName string           `json:"packageName"` // The name of the package
	Path        string           `json:"path"`        // The path to the file, relative to nexema.yaml
	Types       []TypeDefinition `json:"types"`       // The list of types defined
}

// NexemaSnapshot represents a generated project definition
type NexemaSnapshot struct {
	Version  int          `json:"version"` // The Nexema version, at the moment, always 1
	Hashcode string       `json:"hashcode"`
	Files    []NexemaFile `json:"files"`
}

func (s *NexemaSnapshot) FindFile(id string) *NexemaFile {
	for _, file := range s.Files {
		if file.Id == id {
			return &file
		}
	}

	return nil
}
