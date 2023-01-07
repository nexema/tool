package internal

import "fmt"

// Analyzer takes an array of Ast and do validations in order to check if it matches the
// Nexema specification
type Analyzer struct {
	astArr []*Ast
}

// NewAnalyzer creates a new Analzer
func NewAnalyzer(input []*Ast) *Analyzer {
	return &Analyzer{
		astArr: input,
	}
}

// AnalyzeSyntax is the first step of the Analyzer, which analyzes the given list of Ast when creating it,
// checking if they match the Nexema's definition. It does not build nothing. Imports are not checked here.
// It reports any error that is encountered.
func (a *Analyzer) AnalyzeSyntax() error {
	for _, ast := range a.astArr {
		err := a.validateAst(ast)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Analyzer) validateAst(ast *Ast) error {
	for _, typeStmt := range *ast.types {
		err := a.validateType(typeStmt)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateType values a TypeStmt against the following rules:
// 1- Metadata validates successfully
// 2- Fields index are unique and correlative starting from 0 if Type is Enum.
// 3- Each field validates successfully
func (a *Analyzer) validateType(stmt *TypeStmt) error {

	// rule 1
	if stmt.metadata != nil {
		a.validateMetadata(stmt.metadata)
	}

	// rule 2
	if len(*stmt.fields) > 0 {

		// todo: a lot of optimizations can be done
		if stmt.modifier == Token_Enum {
			lastIdx := int64(0)
			if (*stmt.fields)[0].index != nil {
				lastIdx = ((*stmt.fields)[0].index.(*PrimitiveValueStmt)).value.(int64)
				if lastIdx != 0 {
					return a.err("expected the first field in a enum to have the index 0, given index %d", lastIdx)
				}
			}

			usedIndexes := map[int64]bool{}
			for i, field := range *stmt.fields {
				if i == 0 { // skip because it was checked before the loop
					continue
				}

				if field.index == nil {
					// assign the field index
					field.index = &PrimitiveValueStmt{
						value: lastIdx + 1,
						kind:  Primitive_Int64,
					}
					lastIdx++
					usedIndexes[lastIdx] = true
				} else {
					fieldIndex := (field.index.(*PrimitiveValueStmt)).value.(int64)

					// check uniqueness
					_, used := usedIndexes[fieldIndex]
					if used {
						return a.err("index %d already defined for a field", fieldIndex)
					}

					// verify its correlative
					if lastIdx+1 == fieldIndex {
						lastIdx = fieldIndex
					} else {
						return a.err("field indexes in a enum must be correlative")
					}
				}

				// rule 3
				err := a.validateField(field)
				if err != nil {
					return err
				}
			}
		} else {
			lastIdx := int64(-1)
			usedIndexes := map[int64]bool{}
			for _, field := range *stmt.fields {
				if field.index == nil {
					// assign the field index
					field.index = &PrimitiveValueStmt{
						value: lastIdx + 1,
						kind:  Primitive_Int64,
					}
					lastIdx++
					usedIndexes[lastIdx] = true
				} else {
					fieldIndex := (field.index.(*PrimitiveValueStmt)).value.(int64)
					_, used := usedIndexes[fieldIndex]
					if used {
						return a.err("index %d already defined for a field", fieldIndex)
					}

					lastIdx = fieldIndex
				}

				// rule 3
				err := a.validateField(field)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// validateField validates a FieldStmt against the following rules:
// 1- Default value type matches field's type
// 2- Metadata validates successfully
//
// It does not validates imports or custom types
func (a *Analyzer) validateField(f *FieldStmt) error {
	// rule 1
	if f.defaultValue != nil {
		// todo: verify value
	}

	return nil
}

// validateMetadata validates a MapValueStmt against the following rules:
// 1- Map keys are of type string
// 2- Map values are one of the following: string|bool|float64|int64
// 3- There are no duplicated keys
func (a *Analyzer) validateMetadata(m *MapValueStmt) error {
	for _, entry := range *m {
		// validate rule 1
		if entry.key.Kind() != Primitive_String {
			return a.err("metadata map keys must be of type string")
		}

		// validate rule 2
		switch entry.value.Kind() {
		case Primitive_String, Primitive_Bool, Primitive_Float64, Primitive_Int64:
			continue

		default:
			return a.err("metadata map values must be one of the following types: string|bool|float64|int64, given: %s", entry.value.Kind().String())
		}
	}

	// Validate rule 3
	// todo: it will iterate again over the complete map, and maybe can be improved
	return a.validateMap(m)
}

// validateMap validates a MapValueStmt against the following rules:
// 1- Key is not a list, map nor type except enum
// 2- There are no duplicated keys
func (a *Analyzer) validateMap(m *MapValueStmt) error {
	keys := map[any]any{} // map to check for duplicated keys

	for _, entry := range *m {
		key := entry.key

		// check rule 1
		switch key.Kind() {
		case Primitive_List, Primitive_Map, Primitive_Null:
			return a.err("map keys cannot be of type list, map nor null")
		}

		// check rule 2
		raw := (key.(*PrimitiveValueStmt)).value
		_, used := keys[raw]
		if used {
			return a.err("key %v already exists in map", raw)
		}
	}

	return nil
}

func (a *Analyzer) err(txt string, args ...any) error {
	if len(args) > 0 {
		return fmt.Errorf("[analyzer] 0:0 -> %s", args...)
	}
	return fmt.Errorf("[analyzer] 0:0 -> %s", txt)
}
