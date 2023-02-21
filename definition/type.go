package definition

import "tomasweigenast.com/nexema/tool/token"

// TypeDefinition represents a Nexema's type
type TypeDefinition struct {
	Id            uint64
	Name          string
	Documentation []string
	Annotations   Assignments
	Modifier      token.TokenKind
	BaseType      *uint64
	Fields        []*FieldDefinition
}
