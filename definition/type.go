package definition

import "tomasweigenast.com/nexema/tool/token"

// TypeDefinition represents a Nexema's type
type TypeDefinition struct {
	Id            string             `json:"id"`
	Name          string             `json:"name"`
	Documentation []string           `json:"documentation"`
	Annotations   Assignments        `json:"annotations"`
	Modifier      token.TokenKind    `json:"modifier"`
	BaseType      *string            `json:"baseType"`
	Fields        []*FieldDefinition `json:"fields"`
	Defaults      Assignments        `json:"defaults"`
}
