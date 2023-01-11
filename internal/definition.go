package internal

// NexemaDefinition represents an analyzed and built list of Ast.
// This type is next sent to a plugin to generate source code.
type NexemaDefinition struct {
	Version  int          `json:"version"`  // The Nexema specification's version used to build this definition
	Hashcode uint64       `json:"hashcode"` // Hashcode of the current generation
	Files    []NexemaFile `json:"files"`    // A list of nexema files
}

// NexemaFile represents a .nex file with many NexemaTypeDefinition's
type NexemaFile struct {
	Name  string                 `json:"name"`  // The relative path to the file. Its relative to nexema.yaml
	Types []NexemaTypeDefinition `json:"types"` // The list of types declared in this file
}

// NexemaTypeDefinition contains information about a parsed Nexema type
type NexemaTypeDefinition struct {
	Id            string                      `json:"id"`            // An id generated for this type. It's: sha256(NexemaFilePath-TypeName)
	Name          string                      `json:"name"`          // The name of the type
	Modifier      string                      `json:"modifier"`      // The type's modifier
	Documentation []string                    `json:"documentation"` // The documentation for the type
	Fields        []NexemaTypeFieldDefinition `json:"fields"`        // The list of fields declared in this type
}

// NexemaTypeFieldDefinition contains information about a field declared in a Nexema type
type NexemaTypeFieldDefinition struct {
	Index        int64           `json:"index"`        // The field's index
	Name         string          `json:"name"`         // The field's name
	Metadata     map[string]any  `json:"metadata"`     // The field's metadata
	DefaultValue any             `json:"defaultValue"` // The field's default value
	Type         NexemaValueType `json:"type"`         // The field's value type
}

// BaseNexemaValueType is a base struct for every Nexema's type
type BaseNexemaValueType struct {
	Kind     string `json:"$type"`    // NexemaPrimitiveValueType or NexemaTypeValueType
	Nullable bool   `json:"nullable"` // True if the type is nullable
}

type NexemaValueType interface {
	t() // just to allow NexemaPrimitiveValueType and NexemaTypeValueType be part of this
}

// NexemaPrimitiveValueType represents the value type of a NexemaTypeFieldDefinition
// which has a primitive type.
type NexemaPrimitiveValueType struct {
	Base          BaseNexemaValueType `json:",inline"`
	Primitive     string              `json:"primitive"`     // Value's type primitive
	TypeArguments []NexemaValueType   `json:"typeArguments"` // Any generic type argument
}

// NexemaTypeValueType represents the value type of a NexemaTypeFieldDefinition
// which has another Nexema type as value type.
type NexemaTypeValueType struct {
	Base        BaseNexemaValueType `json:",inline"`
	TypeId      string              `json:"typeId"`      // The imported type's id
	ImportAlias *string             `json:"importAlias"` // the import alias, if specified
}

func (NexemaPrimitiveValueType) t() {}
func (NexemaTypeValueType) t()      {}
