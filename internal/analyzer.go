package internal

import (
	"errors"
	"fmt"
)

// Analyzer takes a list of ResolvedContext and do validations in order to check if it matches the
// Nexema specification. If there is no error, a NexemaDefinition is returned.
type Analyzer struct {
	contexts       []*ResolvedContext // the list of contexts given as input
	currentContext *ResolvedContext   // the context that is being validated
	errors         *ErrorCollection   // any error encountered

	typesId map[*TypeStmt]string // a map that contains the id of a TypeStmt

	skipFields bool // test only
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(input []*ResolvedContext) *Analyzer {
	return &Analyzer{
		contexts: input,
		errors:   NewErrorCollection(),
		typesId:  make(map[*TypeStmt]string),
	}
}

// AnalyzeSyntax is the first step of the Analyzer, which analyzes the given list of ResolvedContext when creating it,
// checking if they match the Nexema's definition. It does not build nothing.
// It reports any error that is encountered.
func (a *Analyzer) AnalyzeSyntax() {
	for _, context := range a.contexts {
		a.validateContext(context)
	}
}

// validateContext validates the given Ast in the context against Ast rules
func (a *Analyzer) validateContext(context *ResolvedContext) {
	a.currentContext = context
	a.validateAst(context.owner)
}

// validateAst validates the given ast against each type's rules.
func (a *Analyzer) validateAst(ast *Ast) {
	for _, typeStmt := range *ast.types {
		a.validateType(typeStmt)
	}
}

// validateType values a TypeStmt against the following rules:
// 1- Metadata validates successfully
// 2- Fields index are unique integers and correlative starting from 0 if Type is Enum.
// 3- Each field validates successfully
func (a *Analyzer) validateType(stmt *TypeStmt) {

	id := HashString(fmt.Sprintf("%s_%s", a.currentContext.owner.file.pkg, stmt.name.lit))
	a.typesId[stmt] = id

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
					a.err("expected the first field in an enum to have the index 0, given index %d", lastIdx)
					return
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
					if field.index.Kind() != Primitive_Int64 {
						a.err("field's index must be a number")
						continue
					}

					fieldIndex := (field.index.(*PrimitiveValueStmt)).value.(int64)

					// check uniqueness
					_, used := usedIndexes[fieldIndex]
					if used {
						a.err("index %d already defined for a field", fieldIndex)
						return
					} else {
						usedIndexes[fieldIndex] = true
					}

					// verify its correlative
					if lastIdx+1 == fieldIndex {
						lastIdx = fieldIndex
					} else {
						a.err("field indexes in an enum must be correlative")
						return
					}
				}

				// rule 3
				if !a.skipFields {
					a.validateField(field)
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
					if field.index.Kind() != Primitive_Int64 {
						a.err("field's index must be a number")
						continue
					}

					fieldIndex := (field.index.(*PrimitiveValueStmt)).value.(int64)
					_, used := usedIndexes[fieldIndex]
					if used {
						a.err("index %d already defined for a field", fieldIndex)
						return
					} else {
						usedIndexes[fieldIndex] = true
					}

					lastIdx = fieldIndex
				}

				// rule 3
				if !a.skipFields {
					a.validateField(field)
				}
			}
		}
	}
}

// validateField validates a FieldStmt against the following rules:
// 1- Validate value type
//
//	a) if its a list:
//		- it contains only one type argument
//		- the type argument is not another list or a map
//	b) if its a map:
//		- it contains exactly two type arguments
//		- key and value cannot be another map or a list
//		- key is not nullable and its not a custom type
//	c) if its a custom type:
//		- it exists and can be imported
//
// 2- Default value type matches field's type
// 3- Metadata validates successfully
//
// It does not validates imports or custom types
func (a *Analyzer) validateField(f *FieldStmt) {
	// rule 1
	primitive := GetPrimitive(f.valueType.ident.lit)

	// if its illegal is because its a custom type
	if primitive == Primitive_Illegal || primitive == Primitive_Type {
		// rule 1.c
		var alias *string
		if f.valueType.ident.alias != "" {
			alias = &f.valueType.ident.alias
		}

		_, err := a.currentContext.lookupType(f.valueType.ident.lit, alias)
		if err != nil {
			if errors.Is(err, ErrNeedAlias) {
				a.err("%s already declared, try defining an alias for your import", f.valueType.ident.lit)
				return
			} else {
				a.err("%s not defined, maybe you missed an import?", f.valueType.ident.lit)
			}
		}
	} else if primitive == Primitive_Map {
		// rule 1.b
		if len(*f.valueType.typeArguments) != 2 {
			a.err("map expects exactly two type arguments")
			return
		}

		key := (*f.valueType.typeArguments)[0]

		if key.nullable {
			a.err("map's key type cannot be nullable")
			return
		}

		keyPrimitive := GetPrimitive(key.ident.lit)
		if keyPrimitive == Primitive_Illegal || keyPrimitive == Primitive_List || keyPrimitive == Primitive_Map {
			a.err("map's key type cannot be another list, map or a custom type")
			return
		}

		valuePrimitive := GetPrimitive((*f.valueType.typeArguments)[0].ident.lit)
		if valuePrimitive == Primitive_List || valuePrimitive == Primitive_Map {
			a.err("map's value type cannot be another list or map")
			return
		}
	} else if primitive == Primitive_List {
		// rule 1.a
		if len(*f.valueType.typeArguments) != 1 {
			a.err("list expects exactly one type argument")
			return
		}

		typeArgumentPrimitive := GetPrimitive((*f.valueType.typeArguments)[0].ident.lit)
		if typeArgumentPrimitive == Primitive_List || typeArgumentPrimitive == Primitive_Map {
			a.err("list's value type cannot or type list or map")
			return
		}
	}

	// rule 2
	if f.defaultValue != nil {
		// todo: verify value
		defaultValuePrimitive := f.defaultValue.Kind()
		if primitive != defaultValuePrimitive {
			a.err("field's default value is not of type %s, it is %s", primitive.String(), defaultValuePrimitive.String())
			return
		}

		// do not check if defaultValuePrimitive is Primitive_List because its already checked before
		if primitive == Primitive_List {
			fieldTypeArgumentPrimitive := GetPrimitive((*f.valueType.typeArguments)[0].ident.lit)

			list := f.defaultValue.(*ListValueStmt)
			for _, elem := range *list {
				elemPrimitive := elem.Kind()
				if elemPrimitive != fieldTypeArgumentPrimitive {
					a.err("element from list in default value is not of type %s, it is %s", fieldTypeArgumentPrimitive.String(), elemPrimitive.String())
					continue
				}
			}
		} else if primitive == Primitive_Map {
			fieldKeyTypeArgumentPrimitive := GetPrimitive((*f.valueType.typeArguments)[0].ident.lit)
			fieldValueTypeArgumentPrimitive := GetPrimitive((*f.valueType.typeArguments)[1].ident.lit)

			m := f.defaultValue.(*MapValueStmt)
			for _, entry := range *m {
				keyPrimitive := entry.key.Kind()
				valuePrimitive := entry.value.Kind()

				if keyPrimitive != fieldKeyTypeArgumentPrimitive {
					a.err("entry's key from map in default value is not of type %s, it is %s", fieldKeyTypeArgumentPrimitive.String(), keyPrimitive.String())
					continue
				}

				if valuePrimitive != fieldValueTypeArgumentPrimitive {
					a.err("entry's value from map in default value is not of type %s, it is %s", fieldValueTypeArgumentPrimitive.String(), valuePrimitive.String())
					continue
				}
			}
		}
	}

	// rule 3
	if f.metadata != nil {
		a.validateMap(f.metadata)
	}
}

// validateMetadata validates a MapValueStmt against the following rules:
// 1- Map's keys are of type string
// 2- Map's values are one of the following: string|bool|float64|int64
// 3- Map validates successfully
func (a *Analyzer) validateMetadata(m *MapValueStmt) {
	for _, entry := range *m {
		// validate rule 1
		if entry.key.Kind() != Primitive_String {
			a.err("metadata map keys must be of type string")
			return
		}

		// validate rule 2
		switch entry.value.Kind() {
		case Primitive_String, Primitive_Bool, Primitive_Float64, Primitive_Int64:
			continue

		default:
			a.err("metadata map values must be one of the following types: string|bool|float64|int64, given: %s", entry.value.Kind().String())
			return
		}
	}

	// Validate rule 3
	// todo: it will iterate again over the complete map, and maybe can be improved
	a.validateMap(m)
}

// validateMap validates a MapValueStmt against the following rules:
// 1- Key is not a list, map or type except enum
// 2- There are no duplicated keys
func (a *Analyzer) validateMap(m *MapValueStmt) {
	keys := map[any]any{} // map to check for duplicated keys

	for _, entry := range *m {
		key := entry.key

		// check rule 1
		switch key.Kind() {
		case Primitive_List, Primitive_Map, Primitive_Illegal, Primitive_Null:
			a.err("map's keys cannot be of type list, map, null or a custom type")
			continue
		}

		// check rule 2
		var raw interface{}
		if key.Kind() == Primitive_Type {
			typeValue := (key.(*TypeValueStmt))
			if typeValue.typeName.alias != "" {
				raw = fmt.Sprintf("%s.%s.%s", typeValue.typeName.alias, typeValue.typeName.lit, typeValue.value.lit)
			} else {
				raw = fmt.Sprintf("%s.%s", typeValue.typeName.lit, typeValue.value.lit)
			}
		} else {
			raw = (key.(*PrimitiveValueStmt)).value
		}

		_, used := keys[raw]
		if used {
			a.err(`key "%v" already exists in map`, raw)
			continue
		} else {
			keys[raw] = 1
		}
	}
}

func (a *Analyzer) err(txt string, args ...any) {
	if len(args) > 0 {
		str := fmt.Sprintf(txt, args...)
		a.errors.Report(fmt.Errorf("[analyzer] 0:0 -> %s", str))
	} else {
		a.errors.Report(fmt.Errorf("[analyzer] 0:0 -> %s", txt))
	}
}
