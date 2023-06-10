package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/analyzer"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
	"tomasweigenast.com/nexema/tool/internal/utils"
)

func TestRule_DefaultValueValidField(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "field exists",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.
					NewFieldBuilder("a").
					BasicValueType("string", false).
					Result()).
				Default("a", "hello").
				Result(),
			wantErr: nil,
		},
		{
			name: "field not found",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.
					NewFieldBuilder("a").
					BasicValueType("string", false).
					Result()).
				Default("b", "hello").
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDefaultValueValidField{FieldName: "b"},
			},
		},
		{
			name: "multiple fields",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.
					NewFieldBuilder("a").
					BasicValueType("string", false).
					Result()).
				Field(utils.
					NewFieldBuilder("b").
					BasicValueType("string", false).
					Result()).
				Default("b", "hello").
				Default("c", "holla").
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDefaultValueValidField{FieldName: "c"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &DefaultValueValidField{}
			obj := scope.NewObject(*test.input)
			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_UniqueDefaultValue(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "no duplicated default values",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Default("a", "hello").
				Default("b", true).
				Default("d", int64(25)).
				Result(),
			wantErr: nil,
		},
		{
			name: "one duplicated default value",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Default("c", "hello").
				Default("a", "hello").
				Default("a", true).
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDuplicatedDefaultValue{FieldName: "a"},
			},
		},
		{
			name: "multiple duplicated default values",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Default("c", "hello").
				Default("a", "hello").
				Default("a", true).
				Default("b", true).
				Default("b", float64(5.5)).
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDuplicatedDefaultValue{FieldName: "a"},
				errDuplicatedDefaultValue{FieldName: "b"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &UniqueDefaultValue{}
			obj := scope.NewObject(*test.input)
			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_UniqueFieldName(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "no duplicated fields",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Result()).
				Field(utils.NewFieldBuilder("b").Result()).
				Field(utils.NewFieldBuilder("c").Result()).
				Result(),
			wantErr: nil,
		},
		{
			name: "one duplicate field",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Result()).
				Field(utils.NewFieldBuilder("b").Result()).
				Field(utils.NewFieldBuilder("a").Result()).
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDuplicatedFieldName{FieldName: "a"},
			},
		},
		{
			name: "multiple duplicated fields",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Result()).
				Field(utils.NewFieldBuilder("b").Result()).
				Field(utils.NewFieldBuilder("a").Result()).
				Field(utils.NewFieldBuilder("c").Result()).
				Field(utils.NewFieldBuilder("c").Result()).
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDuplicatedFieldName{FieldName: "a"},
				errDuplicatedFieldName{FieldName: "c"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &UniqueFieldName{}
			obj := scope.NewObject(*test.input)
			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_ValidBaseType(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   []*parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid Base type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Base("Target").
					Result(),

				utils.NewTypeBuilder("Target").
					Modifier(token.Base).
					Result(),
			},
			wantErr: nil,
		},
		{
			name: "invalid Base type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Base("Target").
					Result(),

				utils.NewTypeBuilder("Target").
					Modifier(token.Enum).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				errWrongBaseType{TypeName: "Target"},
			},
		},
		{
			name: "invalid Base type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Base("Target").
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				analyzer.ErrTypeNotFound{Name: "Target"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &ValidBaseType{}
			objs := map[string]*scope.Object{}
			for _, stmt := range test.input {
				obj := scope.NewObject(*stmt)
				objs[obj.Name] = obj
			}

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), objs))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_UniqueFieldIndex(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "unique field indexes",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Index(0).Result()).
				Field(utils.NewFieldBuilder("b").Index(1).Result()).
				Field(utils.NewFieldBuilder("c").Index(2).Result()).
				Field(utils.NewFieldBuilder("d").Result()).
				Field(utils.NewFieldBuilder("e").Index(4).Result()).
				Result(),
			wantErr: nil,
		},
		{
			name: "non unique field indexes",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Index(0).Result()).
				Field(utils.NewFieldBuilder("b").Index(1).Result()).
				Field(utils.NewFieldBuilder("c").Index(1).Result()).
				Field(utils.NewFieldBuilder("d").Result()).
				Field(utils.NewFieldBuilder("e").Index(4).Result()).
				Result(),
			wantErr: []analyzer.AnalyzerErrorKind{
				errDuplicatedFieldIndex{FieldIndex: 1},
			},
		},
		{
			name: "unique field indexes without defining",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").Result()).
				Field(utils.NewFieldBuilder("b").Result()).
				Field(utils.NewFieldBuilder("c").Result()).
				Field(utils.NewFieldBuilder("d").Result()).
				Field(utils.NewFieldBuilder("e").Result()).
				Result(),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &UniqueFieldIndex{}
			obj := scope.NewObject(*test.input)

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_ValidFieldType(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   []*parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid nexema fields",
			input: []*parser.TypeStmt{utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").BasicValueType("string", false).Result()).
				Field(utils.NewFieldBuilder("b").BasicValueType("bool", false).Result()).
				Field(utils.NewFieldBuilder("c").BasicValueType("varint", false).Result()).
				Field(utils.NewFieldBuilder("d").BasicValueType("uvarint", false).Result()).
				Field(utils.NewFieldBuilder("e").BasicValueType("int8", false).Result()).
				Field(utils.NewFieldBuilder("f").BasicValueType("int16", false).Result()).
				Field(utils.NewFieldBuilder("g").BasicValueType("int32", false).Result()).
				Field(utils.NewFieldBuilder("h").BasicValueType("int64", false).Result()).
				Field(utils.NewFieldBuilder("i").BasicValueType("uint8", false).Result()).
				Field(utils.NewFieldBuilder("j").BasicValueType("uint16", false).Result()).
				Field(utils.NewFieldBuilder("k").BasicValueType("uint32", false).Result()).
				Field(utils.NewFieldBuilder("l").BasicValueType("uint64", false).Result()).
				Field(utils.NewFieldBuilder("m").BasicValueType("float32", false).Result()).
				Field(utils.NewFieldBuilder("n").BasicValueType("float64", false).Result()).
				Field(utils.NewFieldBuilder("o").BasicValueType("timestamp", false).Result()).
				Field(utils.NewFieldBuilder("p").BasicValueType("duration", false).Result()).
				Result()},
			wantErr: nil,
		},
		{
			name: "valid custom value type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").BasicValueType("Other", false).Result()).
					Result(),

				utils.NewTypeBuilder("Other").
					Modifier(token.Enum).
					Result(),
			},
			wantErr: nil,
		},
		{
			name: "unknown value type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").BasicValueType("Other", false).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				analyzer.ErrTypeNotFound{Name: "Other"},
			},
		},
		{
			name: "circular import",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").BasicValueType("Test", false).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				errTypeNotAllowed{},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &ValidFieldType{}
			objs := map[string]*scope.Object{}
			for _, stmt := range test.input {
				obj := scope.NewObject(*stmt)
				objs[obj.Name] = obj
			}

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), objs))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_ValidListArguments(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   []*parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid list field",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("list", "", []string{"string"}, false)).Result()).
					Field(utils.NewFieldBuilder("b").ValueType(utils.NewDeclStmt("list", "", []string{"Other"}, false)).Result()).
					Result(),

				utils.NewTypeBuilder("Other").
					Modifier(token.Enum).
					Result(),
			},
			wantErr: nil,
		},
		{
			name: "invalid length",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("list", "", []string{"string", "bool"}, false)).Result()).
					Result(),
				utils.NewTypeBuilder("Test2").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("list", "", []string{}, false)).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				errInvalidListArgumentsLen{Given: 2},
				errInvalidListArgumentsLen{Given: 0},
			},
		},
		{
			name: "invalid argument type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("list", "", []string{"Unknown"}, false)).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				analyzer.ErrTypeNotFound{Name: "Unknown"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &ValidListArguments{}
			objs := map[string]*scope.Object{}
			for _, stmt := range test.input {
				obj := scope.NewObject(*stmt)
				objs[obj.Name] = obj
			}

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), objs))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_ValidMapArguments(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   []*parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid map field",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "string"}, false)).Result()).
					Field(utils.NewFieldBuilder("b").ValueType(utils.NewDeclStmt("map", "", []string{"string", "Other"}, false)).Result()).
					Result(),

				utils.NewTypeBuilder("Other").
					Modifier(token.Enum).
					Result(),
			},
			wantErr: nil,
		},
		{
			name: "invalid length",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string"}, false)).Result()).
					Result(),
				utils.NewTypeBuilder("Test2").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{}, false)).Result()).
					Result(),
				utils.NewTypeBuilder("Test3").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "bool", "varint"}, false)).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				errInvalidMapArgumentsLen{Given: 1},
				errInvalidMapArgumentsLen{Given: 0},
				errInvalidMapArgumentsLen{Given: 3},
			},
		},
		{
			name: "invalid value argument type",
			input: []*parser.TypeStmt{
				utils.NewTypeBuilder("Test").
					Modifier(token.Struct).
					Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "Unknown"}, false)).Result()).
					Result(),
			},
			wantErr: []analyzer.AnalyzerErrorKind{
				analyzer.ErrTypeNotFound{Name: "Unknown"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &ValidMapArguments{}
			objs := map[string]*scope.Object{}
			for _, stmt := range test.input {
				obj := scope.NewObject(*stmt)
				objs[obj.Name] = obj
			}

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), objs))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_ValidMapKey(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid map key",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"bool", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"uvarint", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"uint8", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"uint16", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"uint32", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"uint64", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"varint", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"int16", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"int32", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"int64", "string"}, false)).Result()).
				Result(),

			wantErr: nil,
		},
		{
			name: "invalid map key",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"float32", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"float64", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"timestamp", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"duration", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"Other", "string"}, false)).Result()).
				Result(),

			wantErr: []analyzer.AnalyzerErrorKind{
				errInvalidMapKey{GivenType: definition.Float32},
				errInvalidMapKey{GivenType: definition.Float64},
				errInvalidMapKey{GivenType: definition.Timestamp},
				errInvalidMapKey{GivenType: definition.Duration},
				errInvalidMapKey{},
			},
		},
		{
			name: "non nullable keys",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Struct).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("string", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("bool", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("uvarint", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("uint8", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("uint16", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("uint32", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("uint64", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("varint", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("int8", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("int16", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("int32", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewFullDeclStmt("map", "", []parser.DeclStmt{*utils.NewSimpleDeclStmt("int64", true), *utils.NewSimpleDeclStmt("string", false)}, false)).Result()).
				Result(),

			wantErr: []analyzer.AnalyzerErrorKind{
				errInvalidMapKey{GivenType: definition.String},
				errInvalidMapKey{GivenType: definition.Boolean},
				errInvalidMapKey{GivenType: definition.Uint},
				errInvalidMapKey{GivenType: definition.Uint8},
				errInvalidMapKey{GivenType: definition.Uint16},
				errInvalidMapKey{GivenType: definition.Uint32},
				errInvalidMapKey{GivenType: definition.Uint64},
				errInvalidMapKey{GivenType: definition.Int},
				errInvalidMapKey{GivenType: definition.Int8},
				errInvalidMapKey{GivenType: definition.Int16},
				errInvalidMapKey{GivenType: definition.Int32},
				errInvalidMapKey{GivenType: definition.Int64},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &ValidMapKey{}
			obj := scope.NewObject(*test.input)

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}

func TestRule_NonNullableUnionField(t *testing.T) {

	for _, test := range []struct {
		name    string
		input   *parser.TypeStmt
		wantErr []analyzer.AnalyzerErrorKind
	}{
		{
			name: "valid union fields",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Union).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "string"}, false)).Result()).
				Field(utils.NewFieldBuilder("b").ValueType(utils.NewDeclStmt("string", "", []string{}, false)).Result()).
				Field(utils.NewFieldBuilder("c").ValueType(utils.NewDeclStmt("bool", "", []string{}, false)).Result()).
				Result(),

			wantErr: nil,
		},
		{
			name: "nullable union fields",
			input: utils.NewTypeBuilder("Test").
				Modifier(token.Union).
				Field(utils.NewFieldBuilder("a").ValueType(utils.NewDeclStmt("map", "", []string{"string", "string"}, true)).Result()).
				Field(utils.NewFieldBuilder("b").ValueType(utils.NewDeclStmt("string", "", []string{}, true)).Result()).
				Field(utils.NewFieldBuilder("c").ValueType(utils.NewDeclStmt("bool", "", []string{}, true)).Result()).
				Result(),

			wantErr: []analyzer.AnalyzerErrorKind{
				errNonNullableUnionField{},
				errNonNullableUnionField{},
				errNonNullableUnionField{},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := &parser.File{Path: "test"}
			rule := &NonNullableUnionField{}
			obj := scope.NewObject(*test.input)

			context := analyzer.NewAnalyzerContext(scope.NewLocalScope(file, make(map[string]*scope.Import), map[string]*scope.Object{
				obj.Name: obj,
			}))

			rule.Analyze(context)
			errors := context.Errors()

			if len(test.wantErr) > 0 && errors.IsEmpty() {
				t.Errorf("expected errors (%#v) but got none", test.wantErr)
			} else if len(test.wantErr) > 0 && !errors.IsEmpty() {
				gotErrors := make([]analyzer.AnalyzerErrorKind, 0)
				errors.Iterate(func(err *analyzer.AnalyzerError) {
					gotErrors = append(gotErrors, err.Kind)
				})

				require.Equal(t, test.wantErr, gotErrors)
			}
		})
	}
}
