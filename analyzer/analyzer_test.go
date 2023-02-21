package analyzer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tidwall/btree"
	"tomasweigenast.com/nexema/tool/definition"
	"tomasweigenast.com/nexema/tool/parser"
	"tomasweigenast.com/nexema/tool/scope"
	"tomasweigenast.com/nexema/tool/token"
	"tomasweigenast.com/nexema/tool/tokenizer"
)

func TestAnalyzer_ValidateField(t *testing.T) {
	tests := []struct {
		name         string
		input        parser.FieldStmt
		inputNames   []string
		inputIndexes []int
		typeModifier token.TokenKind
		wantDef      *definition.FieldDefinition
		wantErrs     *AnalyzerErrorCollection
	}{
		{
			name: "field without index",
			input: parser.FieldStmt{
				Name:      parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "string"), Nullable: false},
			},
			typeModifier: token.Struct,
			wantDef: &definition.FieldDefinition{
				Name: "field_name",
				Type: definition.PrimitiveValueType{Primitive: definition.String, Nullable: false},
			},
			wantErrs: newAnalyzerErrorCollection(),
		},
		{
			name: "field with valid index",
			input: parser.FieldStmt{
				Index:     &parser.IdentStmt{Token: *token.NewToken(token.Integer, "1")},
				Name:      parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "string"), Nullable: false},
			},
			typeModifier: token.Struct,
			wantDef: &definition.FieldDefinition{
				Index: 1,
				Name:  "field_name",
				Type:  definition.PrimitiveValueType{Primitive: definition.String, Nullable: false},
			},
			wantErrs: newAnalyzerErrorCollection(),
		},
		{
			name: "field with duplicated index",
			input: parser.FieldStmt{
				Index:     &parser.IdentStmt{Token: *token.NewToken(token.Integer, "1")},
				Name:      parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "string"), Nullable: false},
			},
			inputIndexes: []int{1},
			typeModifier: token.Struct,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongFieldIndex{Err: ErrBaseWrongFieldIndex_DuplicatedIndex}, *tokenizer.NewPos()),
			},
		},
		{
			name: "enum field that does not start with zero",
			input: parser.FieldStmt{
				Index: &parser.IdentStmt{Token: *token.NewToken(token.Integer, "1")},
				Name:  parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
			},
			typeModifier: token.Enum,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongFieldIndex{Err: ErrBaseWrongFieldIndex_EnumShouldBeZeroBased}, *tokenizer.NewPos()),
			},
		},
		{
			name: "enum field that does start with zero",
			input: parser.FieldStmt{
				Index: &parser.IdentStmt{Token: *token.NewToken(token.Integer, "0")},
				Name:  parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
			},
			typeModifier: token.Enum,
			wantDef: &definition.FieldDefinition{
				Index: 0,
				Name:  "field_name",
			},
			wantErrs: newAnalyzerErrorCollection(),
		},
		{
			name: "enum field that is subsequent with others",
			input: parser.FieldStmt{
				Index: &parser.IdentStmt{Token: *token.NewToken(token.Integer, "3")},
				Name:  parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
			},
			inputIndexes: []int{0, 1, 2},
			typeModifier: token.Enum,
			wantDef: &definition.FieldDefinition{
				Index: 3,
				Name:  "field_name",
			},
			wantErrs: newAnalyzerErrorCollection(),
		},
		{
			name: "enum field that is not subsequent with others",
			input: parser.FieldStmt{
				Index: &parser.IdentStmt{Token: *token.NewToken(token.Integer, "2")},
				Name:  parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
			},
			inputIndexes: []int{0},
			typeModifier: token.Enum,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongFieldIndex{ErrBaseWrongFieldIndex_EnumShouldBeSubsequent}, *tokenizer.NewPos()),
			},
		},
		{
			name: "union cannot declare nullable fields",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token:    *token.NewToken(token.Ident, "string"),
					Nullable: true,
				},
			},
			typeModifier: token.Union,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrNonNullableUnionFields{}, *tokenizer.NewPos()),
			},
		},
		{
			name: "field names cannot be duplicated",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token:    *token.NewToken(token.Ident, "string"),
					Nullable: true,
				},
			},
			inputNames:   []string{"field_name"},
			typeModifier: token.Struct,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrAlreadyDefined{Name: "field_name"}, *tokenizer.NewPos()),
			},
		},
		{
			name: "different field names succeed",
			input: parser.FieldStmt{
				Name:      parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "string")},
			},
			inputNames:   []string{"other_field"},
			typeModifier: token.Struct,
			wantDef: &definition.FieldDefinition{
				Name: "field_name",
				Type: definition.PrimitiveValueType{Primitive: definition.String},
			},
			wantErrs: &AnalyzerErrorCollection{},
		},
		{
			name: "field types are valid and existing types",
			input: parser.FieldStmt{
				Name:      parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "rand")},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrTypeNotFound{Name: "rand"}, *tokenizer.NewPos()),
			},
		},
		{
			name: "list argument is not another list",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "list"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "list"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArguments{Primitive: definition.List}, *tokenizer.NewPos()),
			},
		},
		{
			name: "list expects exactly one argument",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "list"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
						{
							Token: *token.NewToken(token.Ident, "bool"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArgumentsLen{Primitive: definition.List, ArgumentsLen: 2}, *tokenizer.NewPos()),
			},
		},
		{
			name: "map expects exactly two arguments",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "map"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArgumentsLen{Primitive: definition.Map, ArgumentsLen: 1}, *tokenizer.NewPos()),
			},
		},
		{
			name: "map key is not nullable",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "map"),
					Args: []parser.DeclStmt{
						{
							Token:    *token.NewToken(token.Ident, "string"),
							Nullable: true,
						},
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArguments{Primitive: definition.Map, IsMapKey: true}, *tokenizer.NewPos()),
			},
		},
		{
			name: "map key cannot be of type float",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "map"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "float32"),
						},
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArguments{Primitive: definition.Map, IsMapKey: true}, *tokenizer.NewPos()),
			},
		},
		{
			name: "map value cannot be another map",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "map"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
						{
							Token: *token.NewToken(token.Ident, "map"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArguments{Primitive: definition.Map}, *tokenizer.NewPos()),
			},
		},
		{
			name: "map value cannot be a list",
			input: parser.FieldStmt{
				Name: parser.IdentStmt{Token: *token.NewToken(token.Ident, "field_name")},
				ValueType: &parser.DeclStmt{
					Token: *token.NewToken(token.Ident, "map"),
					Args: []parser.DeclStmt{
						{
							Token: *token.NewToken(token.Ident, "string"),
						},
						{
							Token: *token.NewToken(token.Ident, "list"),
						},
					},
				},
			},
			typeModifier: token.Struct,
			wantDef:      nil,
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrWrongArguments{Primitive: definition.Map}, *tokenizer.NewPos()),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*scope.Scope{})
			analyzer.currScope = &scope.Scope{}
			analyzer.currLocalScope = &scope.LocalScope{}
			names := map[string]bool{}
			indexes := btree.Set[int]{}

			for _, name := range test.inputNames {
				names[name] = true
			}

			for _, idx := range test.inputIndexes {
				indexes.Insert(idx)
			}

			gotDef := analyzer.analyzeFieldStmt(&test.input, &names, &indexes, test.typeModifier)
			if diff := cmp.Diff(test.wantErrs, analyzer.errors, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("TestAnalyzer_TestValidateField: %s: wantErrs mismatch (-want +got):\n%s", test.name, diff)
			}

			if analyzer.errors.IsEmpty() {
				if diff := cmp.Diff(test.wantDef, gotDef); diff != "" {
					t.Errorf("TestAnalyzer_TestValidateField: %s: wantDef mismatch (-want +got):\n%s", test.name, diff)
				}
			}
		})
	}
}

func TestAnalyzer_ValidateType(t *testing.T) {
	tests := []struct {
		name     string
		input    parser.TypeStmt
		wantDef  *definition.TypeDefinition
		wantErrs *AnalyzerErrorCollection
	}{
		{
			name: "modifier must be struct, union, base or enum",
			input: parser.TypeStmt{
				Name:     parser.IdentStmt{Token: *token.NewToken(token.Ident, "A")},
				Modifier: token.As,
			},
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrUnknownTypeModifier{Token: token.As}, *tokenizer.NewPos()),
			},
		},
		{
			name: "base struct must exists",
			input: parser.TypeStmt{
				Name:     parser.IdentStmt{Token: *token.NewToken(token.Ident, "A")},
				Modifier: token.Struct,
				BaseType: &parser.DeclStmt{Token: *token.NewToken(token.Ident, "B")},
			},
			wantErrs: &AnalyzerErrorCollection{
				NewAnalyzerError(ErrTypeNotFound{Name: "B"}, *tokenizer.NewPos()),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyzer := NewAnalyzer([]*scope.Scope{})
			analyzer.currScope = &scope.Scope{}
			analyzer.currLocalScope = &scope.LocalScope{}

			gotDef := analyzer.analyzeTypeStmt(&test.input)
			if diff := cmp.Diff(test.wantErrs, analyzer.errors, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("TestAnalyzer_TestValidateType: %s: wantErrs mismatch (-want +got):\n%s", test.name, diff)
			}

			if analyzer.errors.IsEmpty() {
				if diff := cmp.Diff(test.wantDef, gotDef); diff != "" {
					t.Errorf("TestAnalyzer_TestValidateType: %s: wantDef mismatch (-want +got):\n%s", test.name, diff)
				}
			}
		})
	}
}
