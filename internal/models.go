package internal

type CompileResult struct {
	Root             string                  `json:"root" yaml:"root"`             // The folder where the project resides
	RootPackage      string                  `json:"-" yaml:"-"`                   // The name of the root package. Not exported when serializing
	OutputPath       string                  `json:"outputPath" yaml:"outputPath"` // The path where to create generated files
	GeneratorOptions *map[string]interface{} `json:"options" yaml:"options"`       // The list of options passed to the generator
	Declaration      *DeclarationTree        `json:"packages" yaml:"packages"`     // The list of packages as a tree structure
}

type TypeDefinition struct {
	Id       string       `json:"id" yaml:"id"`             // A random id for the type
	Name     string       `json:"name" yaml:"name"`         // The name of the type
	Modifier TypeModifier `json:"modifier" yaml:"modifier"` // The modifier of the type
	Fields   interface{}  `json:"fields" yaml:"fields"`     // The fields in the type
}

type StructTypeField struct {
	Name         string                 `json:"name" yaml:"name"`                 // The name of the field
	Type         FieldTypeValue         `json:"type" yaml:"type"`                 // The field's type
	Index        uint32                 `json:"index" yaml:"index"`               // The index of the field
	DefaultValue interface{}            `json:"defaultValue" yaml:"defaultValue"` // The default value of the field
	Metadata     map[string]interface{} `json:"metadata" yaml:"metadata"`         // Any metadata added to the field
}

type EnumTypeField struct {
	Name  string `json:"name" yaml:"name"`
	Index uint32 `json:"index" yaml:"index"`
}

type FieldTypeValue struct {
	Primitive        FieldTypePrimitive `json:"primitive" yaml:"primitive"`         // The field primitive type
	TypeName         string             `json:"typeName" yaml:"typeName"`           // if Primitive is Custom, this contains the custom type's name
	Nullable         bool               `json:"nullable" yaml:"nullable"`           // A flag that indicates if the field is nullable
	TypeArguments    []FieldTypeValue   `json:"typeArguments" yaml:"typeArguments"` // The field's type arguments, if any
	PackageHierarchy []string           `json:"-" yaml:"-"`                         // The package from where the type is imported, only if its a custom type.
	ImportId         string             `json:"importId" yaml:"importId"`           // The id of the type imported
	ResolveImport    string             `json:"-" yaml:"-"`
}

type TypeDefinitionCollection []*TypeDefinition

type TypeModifier string
type FieldTypePrimitive string

const (
	Struct TypeModifier = "struct"
	Enum   TypeModifier = "enum"
	Union  TypeModifier = "union"
)

const (
	UnknownFieldType FieldTypePrimitive = "unknown"
	Boolean          FieldTypePrimitive = "boolean"
	String           FieldTypePrimitive = "string"
	Uint8            FieldTypePrimitive = "uint8"
	Uint16           FieldTypePrimitive = "uint16"
	Uint32           FieldTypePrimitive = "uint32"
	Uint64           FieldTypePrimitive = "uint64"
	Int8             FieldTypePrimitive = "int8"
	Int16            FieldTypePrimitive = "int16"
	Int32            FieldTypePrimitive = "int32"
	Int64            FieldTypePrimitive = "int64"
	Float32          FieldTypePrimitive = "float32"
	Float64          FieldTypePrimitive = "float64"
	Binary           FieldTypePrimitive = "binary"
	List             FieldTypePrimitive = "list"
	Map              FieldTypePrimitive = "map"
	Custom           FieldTypePrimitive = "custom"
)

func ParseTypeModifier(modifier string) (bool, TypeModifier) {
	switch TypeModifier(modifier) {
	case Struct:
		return true, Struct

	case Union:
		return true, Union

	case Enum:
		return true, Enum
	}

	return false, ""
}

func ParseFieldType(fieldType string) (valid bool, primitive FieldTypePrimitive, nullable bool) {
	fieldTypeLen := len(fieldType)
	lastPos := fieldTypeLen - 1
	lastChar := fieldType[lastPos]
	if lastChar == '?' {
		nullable = true
		fieldType = fieldType[:lastPos]
	}

	switch FieldTypePrimitive(fieldType) {
	case Boolean:
		return true, Boolean, nullable

	case String:
		return true, String, nullable

	case Uint8:
		return true, Uint8, nullable

	case Uint16:
		return true, Uint16, nullable

	case Uint32:
		return true, Uint32, nullable

	case Uint64:
		return true, Uint64, nullable

	case Int8:
		return true, Int8, nullable

	case Int16:
		return true, Int16, nullable

	case Int32:
		return true, Int32, nullable

	case Int64:
		return true, Int64, nullable

	case Float32:
		return true, Float32, nullable

	case Float64:
		return true, Float64, nullable

	case Binary:
		return true, Binary, nullable

	case List:
		return true, List, nullable

	case Map:
		return true, Map, nullable
	}

	return false, UnknownFieldType, nullable
}

func (t *TypeDefinitionCollection) LookupType(name string) (td *TypeDefinition, ok bool) {
	for _, typeDefinition := range *t {
		if typeDefinition.Name == name {
			return typeDefinition, true
		}
	}

	return nil, false
}
