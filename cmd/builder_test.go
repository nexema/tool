package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestBuildDefinition(t *testing.T) {
	var tests = []struct {
		name   string
		input  func() []*internal.ResolvedContext
		expect *internal.NexemaDefinition
	}{
		{
			name: "without imports",
			input: func() []*internal.ResolvedContext {
				return []*internal.ResolvedContext{
					{
						Owner: &internal.Ast{
							File: &internal.File{
								Name: "A",
								Pkg:  ".",
							},
							Types: &[]*internal.TypeStmt{
								{
									Name:     &internal.IdentifierStmt{Lit: "StructA"},
									Modifier: internal.Token_Struct,
									Metadata: internal.GetMapValueStmt(map[any]any{
										"hello": int64(2),
									}),
									Documentation: &[]*internal.CommentStmt{
										{Text: "first comment"},
										{Text: "second comment"},
									},
									Fields: &[]*internal.FieldStmt{
										internal.GetField(1, "field_1", "string", false, nil, nil),
										internal.GetField(2, "field_2", "bool", true, nil, nil),
										internal.GetField(3, "field_3", "int64", false, map[any]any{
											"obsolete": true,
										}, int64(25)),
										{
											Index: &internal.PrimitiveValueStmt{RawValue: int64(4), Primitive: internal.Primitive_Int64},
											Name:  &internal.IdentifierStmt{Lit: "field_4"},
											ValueType: &internal.ValueTypeStmt{
												Ident:    &internal.IdentifierStmt{Lit: "StructB"},
												Nullable: true,
											},
										},
									},
								},
								{
									Name:     &internal.IdentifierStmt{Lit: "StructB"},
									Modifier: internal.Token_Union,
									Fields: &[]*internal.FieldStmt{
										internal.GetField(1, "field_1", "float32", false, nil, nil),
										internal.GetField(2, "field_2", "float32", true, nil, nil),
									},
								},
							},
						},
					},
				}
			},
			expect: &internal.NexemaDefinition{
				Version:  builderVersion,
				Hashcode: 1021558313585570668,
				Files: []internal.NexemaFile{
					{
						Name: "A",
						Types: []internal.NexemaTypeDefinition{
							{
								Name:     "StructA",
								Id:       internal.HashString(".-StructA"),
								Modifier: "struct",
								Documentation: []string{
									"first comment",
									"second comment",
								},
								Fields: []internal.NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index:        2,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: true},
											Primitive:     "bool",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index: 3,
										Name:  "field_3",
										Metadata: map[string]any{
											"obsolete": true,
										},
										DefaultValue: int64(25),
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "int64",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index:        4,
										Name:         "field_4",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaTypeValueType{
											Base:   internal.BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId: internal.HashString(".-StructB"),
										},
									},
								},
							},
							{
								Name:          "StructB",
								Id:            internal.HashString(".-StructB"),
								Modifier:      "union",
								Documentation: []string{},
								Fields: []internal.NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "float32",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index:        2,
										Name:         "field_2",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: true},
											Primitive:     "float32",
											TypeArguments: []internal.NexemaValueType{},
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
			input: func() []*internal.ResolvedContext {
				astA := &internal.Ast{
					File: &internal.File{
						Name: "A",
						Pkg:  "a",
					},
					Imports: &[]*internal.ImportStmt{
						{
							Path: &internal.IdentifierStmt{Lit: "b"},
						},
					},
					Types: &[]*internal.TypeStmt{
						{
							Name:          &internal.IdentifierStmt{Lit: "StructA"},
							Modifier:      internal.Token_Struct,
							Metadata:      &internal.MapValueStmt{},
							Documentation: &[]*internal.CommentStmt{},
							Fields: &[]*internal.FieldStmt{
								internal.GetField(1, "field_1", "string", false, nil, nil),
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(2), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_2"},
									ValueType: &internal.ValueTypeStmt{
										Ident:    &internal.IdentifierStmt{Lit: "EnumB"},
										Nullable: true,
									},
								},
							},
						},
					},
				}

				astB := &internal.Ast{
					File: &internal.File{
						Name: "B",
						Pkg:  "b",
					},
					Types: &[]*internal.TypeStmt{
						{
							Name:          &internal.IdentifierStmt{Lit: "EnumB"},
							Modifier:      internal.Token_Enum,
							Metadata:      &internal.MapValueStmt{},
							Documentation: &[]*internal.CommentStmt{},
							Fields: &[]*internal.FieldStmt{
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(0), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_1"},
								},
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(1), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_2"},
								},
							},
						},
					},
				}

				return []*internal.ResolvedContext{
					{
						Owner: astA,
						Dependencies: map[string][]struct {
							Source *internal.Ast
							Alias  *string
						}{
							"B": {
								struct {
									Source *internal.Ast
									Alias  *string
								}{
									Source: astB,
									Alias:  nil,
								},
							},
						},
					},
					{
						Owner: astB,
					},
				}
			},
			expect: &internal.NexemaDefinition{
				Hashcode: 17243524227133482731,
				Version:  builderVersion,
				Files: []internal.NexemaFile{
					{
						Name: "a/A",
						Types: []internal.NexemaTypeDefinition{
							{
								Name:          "StructA",
								Id:            internal.HashString("a-StructA"),
								Modifier:      "struct",
								Documentation: []string{},
								Fields: []internal.NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index:    2,
										Name:     "field_2",
										Metadata: map[string]any{},
										Type: internal.NexemaTypeValueType{
											Base:   internal.BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId: internal.HashString("b-EnumB"),
										},
									},
								},
							},
						},
					},
					{
						Name: "b/B",
						Types: []internal.NexemaTypeDefinition{
							{
								Name:          "EnumB",
								Id:            internal.HashString("b-EnumB"),
								Modifier:      "enum",
								Documentation: []string{},
								Fields: []internal.NexemaTypeFieldDefinition{
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
			input: func() []*internal.ResolvedContext {
				astA := &internal.Ast{
					File: &internal.File{
						Name: "A",
						Pkg:  "a",
					},
					Imports: &[]*internal.ImportStmt{
						{
							Path: &internal.IdentifierStmt{Lit: "b", Alias: "my_b"},
						},
					},
					Types: &[]*internal.TypeStmt{
						{
							Name:          &internal.IdentifierStmt{Lit: "StructA"},
							Modifier:      internal.Token_Struct,
							Metadata:      &internal.MapValueStmt{},
							Documentation: &[]*internal.CommentStmt{},
							Fields: &[]*internal.FieldStmt{
								internal.GetField(1, "field_1", "string", false, nil, nil),
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(2), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_2"},
									ValueType: &internal.ValueTypeStmt{
										Ident:    &internal.IdentifierStmt{Alias: "my_b", Lit: "EnumB"},
										Nullable: true,
									},
								},
							},
						},
					},
				}

				astB := &internal.Ast{
					File: &internal.File{
						Name: "B",
						Pkg:  "b",
					},
					Types: &[]*internal.TypeStmt{
						{
							Name:          &internal.IdentifierStmt{Lit: "EnumB"},
							Modifier:      internal.Token_Enum,
							Metadata:      &internal.MapValueStmt{},
							Documentation: &[]*internal.CommentStmt{},
							Fields: &[]*internal.FieldStmt{
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(0), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_1"},
								},
								{
									Index: &internal.PrimitiveValueStmt{RawValue: int64(1), Primitive: internal.Primitive_Int64},
									Name:  &internal.IdentifierStmt{Lit: "field_2"},
								},
							},
						},
					},
				}

				return []*internal.ResolvedContext{
					{
						Owner: astA,
						Dependencies: map[string][]struct {
							Source *internal.Ast
							Alias  *string
						}{
							"B": {
								struct {
									Source *internal.Ast
									Alias  *string
								}{
									Source: astB,
									Alias:  internal.String("my_b"),
								},
							},
						},
					},
					{
						Owner: astB,
					},
				}
			},
			expect: &internal.NexemaDefinition{
				Version:  builderVersion,
				Hashcode: 12233624741073429459,
				Files: []internal.NexemaFile{
					{
						Name: "a/A",
						Types: []internal.NexemaTypeDefinition{
							{
								Name:          "StructA",
								Id:            internal.HashString("a-StructA"),
								Modifier:      "struct",
								Documentation: []string{},
								Fields: []internal.NexemaTypeFieldDefinition{
									{
										Index:        1,
										Name:         "field_1",
										Metadata:     map[string]any{},
										DefaultValue: nil,
										Type: internal.NexemaPrimitiveValueType{
											Base:          internal.BaseNexemaValueType{Kind: "NexemaPrimitiveValueType", Nullable: false},
											Primitive:     "string",
											TypeArguments: []internal.NexemaValueType{},
										},
									},
									{
										Index:    2,
										Name:     "field_2",
										Metadata: map[string]any{},
										Type: internal.NexemaTypeValueType{
											Base:        internal.BaseNexemaValueType{Kind: "NexemaTypeValueType", Nullable: true},
											TypeId:      internal.HashString("b-EnumB"),
											ImportAlias: internal.String("my_b"),
										},
									},
								},
							},
						},
					},
					{
						Name: "b/B",
						Types: []internal.NexemaTypeDefinition{
							{
								Name:          "EnumB",
								Id:            internal.HashString("b-EnumB"),
								Modifier:      "enum",
								Documentation: []string{},
								Fields: []internal.NexemaTypeFieldDefinition{
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

			a := internal.NewAnalyzer(tt.input())
			contexts, typesId, errs := a.AnalyzeSyntax()
			require.Empty(t, errs)

			builder := NewBuilder()
			builder.analyzer = a
			builder.contexts = contexts
			builder.typesId = typesId

			def := builder.buildDefinition()
			require.Equal(t, tt.expect, def)
		})
	}
}
