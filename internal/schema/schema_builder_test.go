package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/reference"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestBuilderSnapshot(t *testing.T) {
	root := scope.NewPackageScope("./", nil).(*scope.PackageScope)
	scope0 := scope.NewPackageScope("./common", root).(*scope.PackageScope)
	scope1 := scope.NewPackageScope("./identity", root).(*scope.PackageScope)
	root.Children = append(root.Children, scope0, scope1)

	commonScope := scope.NewFileScope("./common/common.nex", &parser.Ast{}, scope0).(*scope.FileScope)
	commonScope.Objects["Entity"] = scope.NewObject(*utils.NewTypeBuilder("Entity").
		Modifier(token.Base).
		Field(utils.
			NewFieldBuilder("id").
			BasicValueType("string", false).
			Result()).
		Field(utils.
			NewFieldBuilder("created_at").
			BasicValueType("timestamp", false).
			Result()).
		Result())
	scope0.Children = append(scope0.Children, commonScope)

	usersScope := scope.NewFileScope("./identity/users.nex", &parser.Ast{}, scope1).(*scope.FileScope)
	usersScope.Objects["User"] = scope.NewObject(*utils.
		NewTypeBuilder("User").
		Modifier(token.Struct).
		Base("Entity").
		Field(utils.
			NewFieldBuilder("name").
			BasicValueType("string", false).
			Result()).
		Field(utils.
			NewFieldBuilder("phone_number").
			BasicValueType("string", true).
			Result()).
		Result())
	usersScope.Imports.Push("", scope.NewImport("common", "", scope0, *reference.NewPos()))
	scope1.Children = append(scope1.Children, usersScope)

	schemaBuilder := NewSchemaBuilder(root)
	snapshot := schemaBuilder.BuildSnapshot()
	require.Equal(t, definition.NexemaSnapshot{
		Version:  1,
		Hashcode: "7032199189098554956",
		Files: []definition.NexemaFile{
			{
				Id:          "12704812692425306373",
				PackageName: "common",
				Path:        "./common/common.nex",
				Types: []definition.TypeDefinition{
					{
						Id:       "3946789668698579447",
						Name:     "Entity",
						Modifier: token.Base,
						BaseType: nil,
						Fields: []*definition.FieldDefinition{
							{
								Index: 0,
								Name:  "id",
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
							{
								Index: 1,
								Name:  "created_at",
								Type:  definition.PrimitiveValueType{Primitive: definition.Timestamp},
							},
						},
					},
				},
			},
			{
				Id:          "5294643553905539495",
				PackageName: "identity",
				Path:        "./identity/users.nex",
				Types: []definition.TypeDefinition{
					{
						Id:       "7656655605636839541",
						Name:     "User",
						Modifier: token.Struct,
						BaseType: utils.StringPtr("3946789668698579447"),
						Fields: []*definition.FieldDefinition{
							{
								Index: 0,
								Name:  "name",
								Type:  definition.PrimitiveValueType{Primitive: definition.String},
							},
							{
								Index: 1,
								Name:  "phone_number",
								Type:  definition.PrimitiveValueType{Primitive: definition.String, Nullable: true},
							},
						},
					},
				},
			},
		},
	}, *snapshot)
}
