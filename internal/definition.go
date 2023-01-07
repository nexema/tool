package internal

// NexemaDefinition represents an analyzed and built list of Ast.
// This type is next sent to a plugin to generate source code.
type NexemaDefinition struct {
	Version  int          `json:"version"`  // The Nexema specification's version used to build this definition
	Hashcode int          `json:"hashcode"` // Hashcode of the current generation
	Files    []NexemaFile `json:"files"`    // A list of nexema files
}

// NexemaFile represents a .nex file with many NexemaTypeDefinition's
type NexemaFile struct {
	Name  string                 `json:"name"`  // The relative path to the file. Its relative to nexema.yaml
	Types []NexemaTypeDefinition `json:"types"` // The list of types declared in this file
}

// NexemaTypeDefinition contains information about a parsed Nexema type
type NexemaTypeDefinition struct {
	Id            string   `json:"id"`            // An id generated for this type. It's: sha256(NexemaFilePath.TypeName)
	Name          string   `json:"name"`          // The name of the type
	Modifier      string   `json:"modifier"`      // The type's modifier
	Documentation []string `json:"documentation"` // The documentation for the type
}
