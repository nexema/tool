package internal

import (
	"errors"
	"fmt"
)

// Analyzer takes a list of ResolvedContext and do validations in order to check if it matches the
// Nexema specification. If there is no error, a NexemaDefinition is returned.
type Analyzer struct {
	errors              *ErrorCollection // any error encountered
	scopes              *ScopeCollection
	currentPackageScope *PackageScope
	currentLocalScope   *LocalScope
	currentAst          *Ast
	currentObject       *Object

	skipFields bool // test only
}

// NewAnalyzer creates a new Analyzer
func NewAnalyzer(scopes *ScopeCollection) *Analyzer {
	return &Analyzer{
		errors: NewErrorCollection(),
		scopes: scopes,
	}
}

// AnalyzeSyntax is the first step of the Analyzer, which analyzes the given list of ResolvedContext when creating it,
// checking if they match the Nexema's definition. It does not build nothing. It generates ids for each type and store them
// in a.typesId.
// It reports any error that is encountered.
func (a *Analyzer) AnalyzeSyntax() *ErrorCollection {
	// for _, context := range a.contexts {
	// 	a.validateContext(context)
	// }

	for _, scope := range a.scopes.Scopes {
		a.validateScope(scope)
	}

	return a.errors
}

// validateScope validates the current PackageScope
func (a *Analyzer) validateScope(scope *PackageScope) {
	a.currentPackageScope = scope
	for participant, localScope := range scope.Participants {
		a.currentLocalScope = localScope
		a.validateAst(participant)
	}
}

// validateAst validates the given ast against each type's rules.
func (a *Analyzer) validateAst(ast *Ast) {
	a.currentAst = ast
	for _, typeStmt := range *ast.Types {
		a.validateType(typeStmt)
	}
}

// validateType values a TypeStmt against the following rules:
// 1- Metadata validates successfully
// 2- Fields index are unique integers and correlative starting from 0 if Type is Enum.
// 3- Each field validates successfully
// 4- Its not imported itself (a field is not of type [CurrentType])
func (a *Analyzer) validateType(stmt *TypeStmt) {

	a.currentObject = a.currentLocalScope.root.Objects.find(stmt.Name.Lit)

	// rule 1
	if stmt.Metadata != nil {
		a.validateMetadata(stmt.Metadata)
	}

	// rule 2
	if stmt.Fields != nil && len(*stmt.Fields) > 0 {

		// todo: a lot of optimizations can be done
		if stmt.Modifier == Token_Enum {
			lastIdx := int64(0)
			if (*stmt.Fields)[0].Index != nil {
				lastIdx = ((*stmt.Fields)[0].Index.(*PrimitiveValueStmt)).RawValue.(int64)
				if lastIdx != 0 {
					a.err("expected the first field in an enum to have the index 0, given index %d", lastIdx)
					return
				}
			} else {
				(*stmt.Fields)[0].Index = &PrimitiveValueStmt{
					RawValue:  int64(0),
					Primitive: Primitive_Int64,
				}
			}

			usedIndexes := map[int64]bool{}
			for i, field := range *stmt.Fields {
				if i == 0 { // skip because it was checked before the loop
					continue
				}

				if field.Index == nil {
					// assign the field index
					field.Index = &PrimitiveValueStmt{
						RawValue:  lastIdx + 1,
						Primitive: Primitive_Int64,
					}
					lastIdx++
					usedIndexes[lastIdx] = true
				} else {
					if field.Index.Kind() != Primitive_Int64 {
						a.err("field's index must be a number")
						continue
					}

					fieldIndex := (field.Index.(*PrimitiveValueStmt)).RawValue.(int64)

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
					a.validateField(stmt, field, true, false)
				}
			}
		} else {
			lastIdx := int64(-1)
			usedIndexes := map[int64]bool{}
			for _, field := range *stmt.Fields {
				if field.Index == nil {
					// assign the field index
					field.Index = &PrimitiveValueStmt{
						RawValue:  lastIdx + 1,
						Primitive: Primitive_Int64,
					}
					lastIdx++
					usedIndexes[lastIdx] = true
				} else {
					if field.Index.Kind() != Primitive_Int64 {
						a.err("field's index must be a number")
						continue
					}

					fieldIndex := (field.Index.(*PrimitiveValueStmt)).RawValue.(int64)
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
					a.validateField(stmt, field, false, stmt.Modifier == Token_Union)
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
// 4- Union cannot declare nullable fields
// 5- Type cannot be defined itself
//
// It does not validates imports or custom types
func (a *Analyzer) validateField(typeStmt *TypeStmt, f *FieldStmt, isEnum, isUnion bool) {
	if !isEnum {
		// rule 1
		primitive := GetPrimitive(f.ValueType.Ident.Lit)

		// if its illegal is because its a custom type
		if primitive == Primitive_Illegal || primitive == Primitive_Type {
			primitive = Primitive_Type

			// rule 1.c
			alias := ""
			if f.ValueType.Ident.Alias != "" {
				alias = f.ValueType.Ident.Alias
			}

			obj, err := a.currentPackageScope.LookupObjectFor(a.currentAst, f.ValueType.Ident.Lit, alias)
			if err != nil {
				if errors.Is(err, ErrNeedAlias) {
					a.err("Type %q already declared, try defining an alias for your import", f.ValueType.Ident.Lit)
					return
				} else {
					a.err("Type %q not defined, maybe you missed an import?", f.ValueType.Ident.Lit)
				}
			}

			// rule 5
			if obj != nil {
				if a.currentObject.ID == obj.ID {
					a.err("%q cannot be declared itself in fields", obj.Stmt.Name.Lit)
				}
			}
		} else if primitive == Primitive_Map {
			// rule 1.b
			if len(*f.ValueType.TypeArguments) != 2 {
				a.err("map expects exactly two type arguments")
				return
			}

			key := (*f.ValueType.TypeArguments)[0]

			if key.Nullable {
				a.err("map's key type cannot be nullable")
				return
			}

			keyPrimitive := GetPrimitive(key.Ident.Lit)
			if keyPrimitive == Primitive_Illegal || keyPrimitive == Primitive_List || keyPrimitive == Primitive_Map {
				a.err("map's key type cannot be another list, map or a custom type")
				return
			}

			valuePrimitive := GetPrimitive((*f.ValueType.TypeArguments)[0].Ident.Lit)
			if valuePrimitive == Primitive_List || valuePrimitive == Primitive_Map {
				a.err("map's value type cannot be another list or map")
				return
			}
		} else if primitive == Primitive_List {
			// rule 1.a
			if len(*f.ValueType.TypeArguments) != 1 {
				a.err("list expects exactly one type argument")
				return
			}

			typeArgumentPrimitive := GetPrimitive((*f.ValueType.TypeArguments)[0].Ident.Lit)
			if typeArgumentPrimitive == Primitive_List || typeArgumentPrimitive == Primitive_Map {
				a.err("list's value type cannot or type list or map")
				return
			}
		}

		// rule 2
		if f.DefaultValue != nil {
			if primitive == Primitive_Type {
				a.err("enum fields cannot declare default values")
				return
			}

			// todo: verify value
			defaultValuePrimitive := f.DefaultValue.Kind()

			if primitive != defaultValuePrimitive {
				a.err("field's default value is not of type %s, it is %s", primitive.String(), defaultValuePrimitive.String())
				return
			}

			// do not check if defaultValuePrimitive is Primitive_List because its already checked before
			if primitive == Primitive_List {
				fieldTypeArgumentPrimitive := GetPrimitive((*f.ValueType.TypeArguments)[0].Ident.Lit)

				list := f.DefaultValue.(*ListValueStmt)
				for _, elem := range *list {
					elemPrimitive := elem.Kind()
					if elemPrimitive != fieldTypeArgumentPrimitive {
						a.err("element from list in default value is not of type %s, it is %s", fieldTypeArgumentPrimitive.String(), elemPrimitive.String())
						continue
					}
				}
			} else if primitive == Primitive_Map {
				fieldKeyTypeArgumentPrimitive := GetPrimitive((*f.ValueType.TypeArguments)[0].Ident.Lit)
				fieldValueTypeArgumentPrimitive := GetPrimitive((*f.ValueType.TypeArguments)[1].Ident.Lit)

				m := f.DefaultValue.(*MapValueStmt)
				for _, entry := range *m {
					keyPrimitive := entry.Key.Kind()
					valuePrimitive := entry.Value.Kind()

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

		// rule 4
		if f.ValueType.Nullable && isUnion {
			a.err("union cannot declare nullable fields")
		}
	}

	// rule 3
	if f.Metadata != nil {
		a.validateMap(f.Metadata)
	}
}

// validateMetadata validates a MapValueStmt against the following rules:
// 1- Map's keys are of type string
// 2- Map's values are one of the following: string|bool|float64|int64
// 3- Map validates successfully
func (a *Analyzer) validateMetadata(m *MapValueStmt) {
	for _, entry := range *m {
		// validate rule 1
		if entry.Key.Kind() != Primitive_String {
			a.err("metadata map keys must be of type string")
			return
		}

		// validate rule 2
		switch entry.Value.Kind() {
		case Primitive_String, Primitive_Bool, Primitive_Float64, Primitive_Int64:
			continue

		default:
			a.err("metadata map values must be one of the following types: string|bool|float64|int64, given: %s", entry.Value.Kind().String())
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
		key := entry.Key

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
			if typeValue.TypeName.Alias != "" {
				raw = fmt.Sprintf("%s.%s.%s", typeValue.TypeName.Alias, typeValue.TypeName.Lit, typeValue.RawValue.Lit)
			} else {
				raw = fmt.Sprintf("%s.%s", typeValue.TypeName.Lit, typeValue.RawValue.Lit)
			}
		} else {
			raw = (key.(*PrimitiveValueStmt)).RawValue
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
		a.errors.Report(fmt.Errorf("[analyzer] %s 0:0 -> %s", a.currentAst.File.Name, str))
	} else {
		a.errors.Report(fmt.Errorf("[analyzer] %s 0:0 -> %s", a.currentAst.File.Name, txt))
	}
}
