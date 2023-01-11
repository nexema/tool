package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildDefinition(t *testing.T) {
	var tests = []struct {
		name   string
		input  func() []*ResolvedContext
		expect *NexemaDefinition
	}{
		{
			name: "without imports",
			input: func() []*ResolvedContext {
				return []*ResolvedContext{
					{
						owner: &Ast{
							file: &File{
								name: "A",
								pkg:  ".",
							},
							types: &[]*TypeStmt{
								{
									name:     &IdentifierStmt{lit: "StructA"},
									modifier: Token_Struct,
									metadata: GetMapValueStmt(map[any]any{
										"hello": int64(2),
									}),
									documentation: &[]*CommentStmt{
										{text: "first comment"},
										{text: "second comment"},
									},
									fields: &[]*FieldStmt{
										GetField(1, "field_1", "string", false, nil, nil),
										GetField(2, "field_2", "bool", true, nil, nil),
										GetField(3, "field_3", "int64", false, map[any]any{
											"obsolete": true,
										}, int64(25)),
										{
											index: &PrimitiveValueStmt{value: int64(4), kind: Primitive_Int64},
											name:  &IdentifierStmt{lit: "field_4"},
											valueType: &ValueTypeStmt{
												ident:    &IdentifierStmt{lit: "StructB"},
												nullable: true,
											},
										},
									},
								},
								{
									name:     &IdentifierStmt{lit: "StructB"},
									modifier: Token_Union,
									fields: &[]*FieldStmt{
										GetField(1, "field_1", "float32", false, nil, nil),
										GetField(2, "field_2", "float32", true, nil, nil),
									},
								},
							},
						},
					},
				}
			},
			expect: &NexemaDefinition{
				Version: builderVersion,
				Files: []NexemaFile{
					{
						Name: "A",
						Types: []NexemaTypeDefinition{
							{
								Name:     "StructA",
								Id:       HashString(".-StructA"),
								Modifier: "struct",
								Documentation: []string{
									"first comment",
									"second comment",
								},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index:        2,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: true},
											Primitive:     "bool",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index: 3,
										Name:  "field_3",
										Metadata: map[string]any{
											"obsolete": true,
										},
										DefaultValue: int64(25),
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "int64",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index:        4,
										Name:         "field_4",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaTypeValueType{
											Base:   BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId: HashString(".-StructB"),
										},
									},
								},
							},
							{
								Name:          "StructB",
								Id:            HashString(".-StructB"),
								Modifier:      "union",
								Documentation: []string{},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "float32",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index:        2,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: true},
											Primitive:     "float32",
											TypeArguments: []nexemaValueType{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with imports and no alias",
			input: func() []*ResolvedContext {
				astA := &Ast{
					file: &File{
						name: "A",
						pkg:  "a",
					},
					imports: &[]*ImportStmt{
						{
							path: &IdentifierStmt{lit: "b"},
						},
					},
					types: &[]*TypeStmt{
						{
							name:          &IdentifierStmt{lit: "StructA"},
							modifier:      Token_Struct,
							metadata:      &MapValueStmt{},
							documentation: &[]*CommentStmt{},
							fields: &[]*FieldStmt{
								GetField(1, "field_1", "string", false, nil, nil),
								{
									index: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_2"},
									valueType: &ValueTypeStmt{
										ident:    &IdentifierStmt{lit: "EnumB"},
										nullable: true,
									},
								},
							},
						},
					},
				}

				astB := &Ast{
					file: &File{
						name: "B",
						pkg:  "b",
					},
					types: &[]*TypeStmt{
						{
							name:          &IdentifierStmt{lit: "EnumB"},
							modifier:      Token_Enum,
							metadata:      &MapValueStmt{},
							documentation: &[]*CommentStmt{},
							fields: &[]*FieldStmt{
								{
									index: &PrimitiveValueStmt{value: int64(0), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_1"},
								},
								{
									index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_2"},
								},
							},
						},
					},
				}

				return []*ResolvedContext{
					{
						owner: astA,
						dependencies: map[string][]struct {
							source *Ast
							alias  *string
						}{
							"B": {
								struct {
									source *Ast
									alias  *string
								}{
									source: astB,
									alias:  nil,
								},
							},
						},
					},
					{
						owner: astB,
					},
				}
			},
			expect: &NexemaDefinition{
				Version: builderVersion,
				Files: []NexemaFile{
					{
						Name: "a/A",
						Types: []NexemaTypeDefinition{
							{
								Name:          "StructA",
								Id:            HashString("a-StructA"),
								Modifier:      "struct",
								Documentation: []string{},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index:    2,
										Name:     "field_2",
										Metadata: map[string]any{},
										Type: NexemaTypeValueType{
											Base:   BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId: HashString("b-EnumB"),
										},
									},
								},
							},
						},
					},
					{
						Name: "b/B",
						Types: []NexemaTypeDefinition{
							{
								Name:          "EnumB",
								Id:            HashString("b-EnumB"),
								Modifier:      "enum",
								Documentation: []string{},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        0,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
									},
									{
										Index:        1,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with imports and alias",
			input: func() []*ResolvedContext {
				astA := &Ast{
					file: &File{
						name: "A",
						pkg:  "a",
					},
					imports: &[]*ImportStmt{
						{
							path: &IdentifierStmt{lit: "b", alias: "my_b"},
						},
					},
					types: &[]*TypeStmt{
						{
							name:          &IdentifierStmt{lit: "StructA"},
							modifier:      Token_Struct,
							metadata:      &MapValueStmt{},
							documentation: &[]*CommentStmt{},
							fields: &[]*FieldStmt{
								GetField(1, "field_1", "string", false, nil, nil),
								{
									index: &PrimitiveValueStmt{value: int64(2), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_2"},
									valueType: &ValueTypeStmt{
										ident:    &IdentifierStmt{alias: "my_b", lit: "EnumB"},
										nullable: true,
									},
								},
							},
						},
					},
				}

				astB := &Ast{
					file: &File{
						name: "B",
						pkg:  "b",
					},
					types: &[]*TypeStmt{
						{
							name:          &IdentifierStmt{lit: "EnumB"},
							modifier:      Token_Enum,
							metadata:      &MapValueStmt{},
							documentation: &[]*CommentStmt{},
							fields: &[]*FieldStmt{
								{
									index: &PrimitiveValueStmt{value: int64(0), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_1"},
								},
								{
									index: &PrimitiveValueStmt{value: int64(1), kind: Primitive_Int64},
									name:  &IdentifierStmt{lit: "field_2"},
								},
							},
						},
					},
				}

				return []*ResolvedContext{
					{
						owner: astA,
						dependencies: map[string][]struct {
							source *Ast
							alias  *string
						}{
							"B": {
								struct {
									source *Ast
									alias  *string
								}{
									source: astB,
									alias:  String("my_b"),
								},
							},
						},
					},
					{
						owner: astB,
					},
				}
			},
			expect: &NexemaDefinition{
				Version: builderVersion,
				Files: []NexemaFile{
					{
						Name: "a/A",
						Types: []NexemaTypeDefinition{
							{
								Name:          "StructA",
								Id:            HashString("a-StructA"),
								Modifier:      "struct",
								Documentation: []string{},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: NexemaPrimitiveValueType{
											Base:          BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []nexemaValueType{},
										},
									},
									{
										Index:    2,
										Name:     "field_2",
										Metadata: map[string]any{},
										Type: NexemaTypeValueType{
											Base:   BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId: HashString("b-EnumB"),
										},
									},
								},
							},
						},
					},
					{
						Name: "b/B",
						Types: []NexemaTypeDefinition{
							{
								Name:          "EnumB",
								Id:            HashString("b-EnumB"),
								Modifier:      "enum",
								Documentation: []string{},
								Fields: []NexemaTypeFieldDefinition{
									{
										Index:        0,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
									},
									{
										Index:        1,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			a := NewAnalyzer(tt.input())
			a.AnalyzeSyntax()
			require.Empty(t, a.errors)

			builder := NewBuilder(a)
			def := builder.buildDefinition()
			require.Equal(t, tt.expect, def)
		})
	}
}
