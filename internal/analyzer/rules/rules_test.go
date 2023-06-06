package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
	"tomasweigenast.com/nexema/tool/internal/analyzer"
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

func TestRule_DuplicatedDefaultValue(t *testing.T) {

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
			rule := &DuplicatedDefaultValue{}
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

func TestRule_DuplicatedFieldName(t *testing.T) {

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
			rule := &DuplicatedFieldName{}
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
