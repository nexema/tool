package analyzer

import (
	"path"
	"strconv"
	"strings"

	"github.com/tidwall/btree"
	"tomasweigenast.com/nexema/tool/definition"
	"tomasweigenast.com/nexema/tool/parser"
	"tomasweigenast.com/nexema/tool/scope"
	"tomasweigenast.com/nexema/tool/token"
)

// Analyzer takes a linked list of built scopes and analyzes them syntactically.
// Also, if analysis succeed, a definition is built
type Analyzer struct {
	scopes []*scope.Scope
	errors *AnalyzerErrorCollection

	currScope      *scope.Scope
	currLocalScope *scope.LocalScope
	currTypeId     uint64
	files          []definition.NexemaFile
}

func NewAnalyzer(scopes []*scope.Scope) *Analyzer {
	return &Analyzer{
		scopes: scopes,
		errors: newAnalyzerErrorCollection(),
		files:  make([]definition.NexemaFile, 0),
	}
}

// Analyze starts analyzing and logs any error encountered
func (self *Analyzer) Analyze() {
	for _, scope := range self.scopes {
		self.analyzeScope(scope)
	}
}

func (self *Analyzer) analyzeScope(s *scope.Scope) {
	self.currScope = s
	for _, localScope := range *s.LocalScopes() {
		self.analyzeLocalScope(localScope)
	}
}

func (self *Analyzer) analyzeLocalScope(ls *scope.LocalScope) {
	self.currLocalScope = ls
	file := ls.File()
	nexFile := definition.NexemaFile{
		Types:       make([]definition.TypeDefinition, 0),
		FileName:    file.FileName,
		Path:        file.Path,
		PackageName: path.Base(file.Path),
	}

	for _, obj := range *ls.Objects() {
		self.currTypeId = obj.Id
		def := self.analyzeTypeStmt(obj.Source())
		if def != nil {
			nexFile.Types = append(nexFile.Types, *def)
		}
	}

	self.files = append(self.files, nexFile)
}

// analyzeTypeStmt analyses a TypeStmt in order to match the following set of rules:
//
// 1- modifier is token.Struct, token.Enum, token.Union or token.Base
// 2- if struct extends another type, check it exists and is a valid Base type
// 3- each field validates against its own rules
// 4- each default value, if any, is declared once and points to a valid field
//
// If succeed, it outputs a valid definition.TypeDefinition
func (self *Analyzer) analyzeTypeStmt(stmt *parser.TypeStmt) *definition.TypeDefinition {
	def := new(definition.TypeDefinition)

	// rule 1
	switch stmt.Modifier {
	case token.Struct, token.Enum, token.Union, token.Base:
		def.Modifier = stmt.Modifier
		break

	default:
		self.errors.push(ErrUnknownTypeModifier{stmt.Modifier}, stmt.Name.Pos)
	}

	// rule 2
	if stmt.BaseType != nil {
		obj := self.getObject(stmt.BaseType)
		if obj != nil {
			def.BaseType = &obj.Id
		}
	}

	// rule 3
	fieldNames := map[string]bool{}     // to validate field's rule 2.
	fieldIndexes := new(btree.Set[int]) // to validate field's rule 3.
	for _, field := range stmt.Fields {
		fieldDef := self.analyzeFieldStmt(&field, &fieldNames, fieldIndexes, stmt.Modifier)
		if fieldDef != nil {
			def.Fields = append(def.Fields, fieldDef)
		}
	}

	// rule 4
	if stmt.Defaults != nil {
		def.Defaults = self.getAssignments(&stmt.Defaults, false)
	}

	if stmt.Documentation != nil {
		def.Documentation = sanitizeComments(&stmt.Documentation)
	}

	if stmt.Annotations != nil {
		annotations := make([]parser.AssignStmt, len(stmt.Annotations))
		for i, annotation := range stmt.Annotations {
			annotations[i] = annotation.Assigment
		}
		def.Annotations = self.getAssignments(&annotations, true)
	}

	return def
}

// analyzeFieldStmt analyses a FieldStmt in order to match the following set of rules:
//
// 1- field names are not duplicated
// 2- field types are valid and defined (at definition.NexemaValueType) Nexema value type or a imported custom type.
// 2a- if field value type is a list, contains exactly one argument and is a valid and defined Nexema value type and is not a list nor a map.
// 2b- if field value type is a map, its key and value are valid and defined Nexema value types,
// its key is a non nullable string, bool or int; and its value is not another map nor a list.
// 3- indexes start from 0 for enums (and must be subsequents) and 1 for other type and there are no duplicated ones
// 4- field value type is not the same as the current type
// 5- unions cannot declare nullable fields
//
// If suceeds, outputs a [definition.FieldDefinition]
func (self *Analyzer) analyzeFieldStmt(field *parser.FieldStmt, names *map[string]bool, indexes *btree.Set[int], typeModifier token.TokenKind) *definition.FieldDefinition {
	def := new(definition.FieldDefinition)

	// rule 1
	fieldName := field.Name.Token.Literal
	if _, ok := (*names)[fieldName]; ok {
		self.errors.push(ErrAlreadyDefined{fieldName}, field.Name.Pos)
	} else {
		(*names)[fieldName] = true // update map for next field
		def.Name = fieldName
	}

	// rule 2
	if typeModifier != token.Enum {
		valueType := self.getValueType(field.ValueType)
		if valueType != nil {
			primitiveValueType, ok := valueType.(definition.PrimitiveValueType)
			if ok {
				if typeModifier == token.Union && primitiveValueType.Nullable {
					self.errors.push(ErrNonNullableUnionFields{}, field.Name.Pos)
				}

				switch primitiveValueType.Primitive {
				//rule 2a
				case definition.List:
					if len(primitiveValueType.Arguments) != 1 {
						self.errors.push(ErrWrongArgumentsLen{definition.List, len(primitiveValueType.Arguments)}, field.ValueType.Pos)
						break // stop because the next check can fail if len(..) is 0
					}

					argument, ok := primitiveValueType.Arguments[0].(definition.PrimitiveValueType)
					if ok && (argument.Primitive == definition.List || argument.Primitive == definition.Map) {
						self.errors.push(ErrWrongArguments{Primitive: definition.List}, field.ValueType.Pos)
					}

				// rule 2b
				case definition.Map:
					if len(primitiveValueType.Arguments) != 2 {
						self.errors.push(ErrWrongArgumentsLen{definition.Map, len(primitiveValueType.Arguments)}, field.ValueType.Pos)
						break
					}

					key, ok := primitiveValueType.Arguments[0].(definition.PrimitiveValueType)
					wrongArgs := !ok

					if ok {
						switch key.Primitive {
						case definition.String,
							definition.Int,
							definition.Uint,
							definition.Int8,
							definition.Int16,
							definition.Int32,
							definition.Int64,
							definition.Uint8,
							definition.Uint16,
							definition.Uint32,
							definition.Uint64:

							if key.Nullable {
								wrongArgs = true
							}
							break

						default:
							wrongArgs = true
						}
					}

					if wrongArgs {
						self.errors.push(ErrWrongArguments{definition.Map, true}, field.ValueType.Pos)
					}

					value, ok := primitiveValueType.Arguments[1].(definition.PrimitiveValueType)
					if ok && (value.Primitive == definition.List || value.Primitive == definition.Map) {
						self.errors.push(ErrWrongArguments{Primitive: definition.Map}, field.ValueType.Pos)
					}
				}
			} else {
				// rule 4
				customTypeId := valueType.(definition.CustomValueType).ObjectId
				if customTypeId == self.currTypeId {
					self.errors.push(ErrIllegalUseCycle{field.ValueType.Token.Literal}, field.ValueType.Pos)
				}
			}

			def.Type = valueType
		}
	}

	// rule 3
	var fieldIndex int
	if field.Index == nil {
		if indexes.Len() > 0 {
			fieldIndex, _ = indexes.GetAt(indexes.Len() - 1)
		}
	} else {
		fieldIndex = toInt(field.Index.Token.Literal) // its sure Literal will be a number
	}

	// check if already in use
	if indexes.Contains(fieldIndex) {
		self.errors.push(ErrWrongFieldIndex{ErrBaseWrongFieldIndex_DuplicatedIndex}, field.Index.Pos)
	} else if typeModifier == token.Enum {
		// check if first index is zero
		if indexes.Len() == 0 && fieldIndex != 0 {
			self.errors.push(ErrWrongFieldIndex{ErrBaseWrongFieldIndex_EnumShouldBeZeroBased}, field.Index.Pos)
		} else {
			// check if subsequent
			previousIndex, ok := indexes.GetAt(indexes.Len() - 1)
			if ok && fieldIndex != previousIndex+1 {
				self.errors.push(ErrWrongFieldIndex{ErrBaseWrongFieldIndex_EnumShouldBeSubsequent}, field.Index.Pos)
			}
		}
	}

	def.Index = fieldIndex

	if field.Documentation != nil {
		def.Documentation = sanitizeComments(&field.Documentation)
	}

	if field.Annotations != nil {
		annotations := make([]parser.AssignStmt, len(field.Annotations))
		for i, annotation := range field.Annotations {
			annotations[i] = annotation.Assigment
		}
		def.Annotations = self.getAssignments(&annotations, true)
	}

	return def
}

// getObject under the hood calls FindOjbect on self.currLocalScope and reports any error if any
func (self *Analyzer) getObject(decl *parser.DeclStmt) *scope.Object {
	name, alias := decl.Format()
	obj, needAlias := self.currLocalScope.FindObject(name, alias)
	if obj == nil {
		if needAlias {
			self.errors.push(ErrNeedAlias{}, decl.Pos)
		} else {
			self.errors.push(ErrTypeNotFound{name, alias}, decl.Pos)
		}
	} else {
		if obj.Source().Modifier != token.Base {
			self.errors.push(ErrNotValidBaseType{name, alias}, decl.Pos)
		} else {
			return obj
		}
	}

	return nil
}

func (self *Analyzer) getValueType(decl *parser.DeclStmt) definition.BaseValueType {
	typeName, _ := decl.Format()
	primitive, valid := definition.ParsePrimitive(typeName)
	if !valid {
		obj := self.getObject(decl)
		if obj != nil {
			return definition.CustomValueType{ObjectId: obj.Id, Nullable: decl.Nullable}
		}
	} else {
		valueType := definition.PrimitiveValueType{
			Primitive: primitive,
			Nullable:  decl.Nullable,
		}

		if len(decl.Args) > 0 {
			valueType.Arguments = make([]definition.BaseValueType, len(decl.Args))
			for i, arg := range decl.Args {
				argType := self.getValueType(&arg)
				if argType != nil {
					valueType.Arguments[i] = argType
				}
			}
		}

		return valueType
	}

	return nil
}

// sanitizeComments returns a []string from a []parser.CommentStmt, and trims every comment
func sanitizeComments(arr *[]parser.CommentStmt) []string {
	out := make([]string, len(*arr))
	for i, e := range *arr {
		out[i] = strings.TrimSpace(e.Token.Literal)
	}

	return out
}

// getAssignments takes a []parser.AssignStmt and outputs a valid definition.Assignments.
// If isAnnotation is true, it will validate if the given assignments values are string, int64, float64 or boolean
func (self *Analyzer) getAssignments(arr *[]parser.AssignStmt, isAnnotation bool) definition.Assignments {
	out := make(definition.Assignments, len(*arr))
	for _, e := range *arr {
		key := e.Left.Token.Literal
		if _, ok := out[key]; ok {
			// already declared
			self.errors.push(ErrAssignmentKeyAlreadyInUse{key}, e.Left.Pos)
			continue
		}

		value := e.Right.Kind.Value()
		if isAnnotation {
			switch value.(type) {
			case string, int64, float64, bool:
				break

			default:
				self.errors.push(ErrWrongAnnotationValue{}, e.Right.Pos)
				continue
			}
		}

		out[key] = value
	}

	return out
}

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}
