package schema

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
)

const (
	SCHEMA_VERSION_1 int = 1
)

// SchemaBuilder converts a already analyzed ast to a valid Nexema definition.
type SchemaBuilder struct {
	scopes []*scope.Scope
}

func NewSchemaBuilder(scopes []*scope.Scope) *SchemaBuilder {
	return &SchemaBuilder{scopes}
}

// BuildSnapshot creates a Nexema Snapshot
func (self *SchemaBuilder) BuildSnapshot() *definition.NexemaSnapshot {
	snapshot := &definition.NexemaSnapshot{
		Version: SCHEMA_VERSION_1,
		Files:   make([]definition.NexemaFile, 0),
	}

	ids := make([]string, 0)
	for _, scope := range self.scopes {
		for _, localScope := range *scope.LocalScopes() {
			file := self.buildFile(localScope)
			snapshot.Files = append(snapshot.Files, *file)
			ids = append(ids, file.Id)
		}
	}

	snapshotHashcode, err := hashstructure.Hash(&ids, hashstructure.FormatV2, &hashstructure.HashOptions{})
	if err != nil {
		panic(fmt.Errorf("could not calculate hash for snapshot. Please report this because this is not expected. Error: %v", err))
	}

	snapshot.Hashcode = fmt.Sprint(snapshotHashcode)

	return snapshot
}

func (self *SchemaBuilder) buildFile(localScope *scope.LocalScope) *definition.NexemaFile {
	physicalFile := localScope.File()
	file := &definition.NexemaFile{
		Path:        physicalFile.Path,
		PackageName: path.Base(path.Dir(physicalFile.Path)),
		Types:       make([]definition.TypeDefinition, len(*localScope.Objects())),
	}

	idx := 0
	for _, obj := range *localScope.Objects() {
		stmt := obj.Source()
		typeDef := definition.TypeDefinition{
			Id:            obj.Id,
			Name:          obj.Name,
			Documentation: getComments(&stmt.Documentation),
			Annotations:   getAnnotations(&stmt.Annotations),
			Defaults:      getDefaults(&stmt.Defaults),
			Modifier:      stmt.Modifier,
			Fields:        make([]*definition.FieldDefinition, len(stmt.Fields)),
		}

		if stmt.BaseType != nil {
			name, alias := stmt.BaseType.Format()
			typeDef.BaseType = &localScope.MustFindObject(name, alias).Id
		}

		fieldIndex := 0
		for i, field := range stmt.Fields {
			if field.Index != nil {
				fieldIndex, _ = strconv.Atoi(field.Index.Token.Literal)
			}

			typeDef.Fields[i] = &definition.FieldDefinition{
				Name:          field.Name.Token.Literal,
				Index:         fieldIndex,
				Documentation: getComments(&field.Documentation),
				Annotations:   getAnnotations(&field.Annotations),
				Type:          getValueType(localScope, field.ValueType),
			}

			fieldIndex++
		}

		file.Types[idx] = typeDef
		idx++
	}

	fileId, err := hashstructure.Hash(&file, hashstructure.FormatV2, nil)
	if err != nil {
		panic(fmt.Errorf("could not hash file. Please report this error because its not expected. Error: %v", err))
	}
	file.Id = fmt.Sprint(fileId)
	return file
}

func getComments(stmts *[]parser.CommentStmt) []string {
	if len(*stmts) == 0 {
		return nil
	}

	array := make([]string, len(*stmts))
	for i, stmt := range *stmts {
		array[i] = strings.TrimSpace(stmt.Token.Literal)
	}

	return array
}

func getAnnotations(stmts *[]parser.AnnotationStmt) definition.Assignments {
	if len(*stmts) == 0 {
		return nil
	}

	assignments := make(definition.Assignments, len(*stmts))
	for _, stmt := range *stmts {
		assignments[stmt.Assigment.Left.Token.Literal] = stmt.Assigment.Right.Kind.Value()
	}
	return assignments
}

func getDefaults(stmts *[]parser.AssignStmt) definition.Assignments {
	if len(*stmts) == 0 {
		return nil
	}

	assignments := make(definition.Assignments, len(*stmts))
	for _, stmt := range *stmts {
		assignments[stmt.Left.Token.Literal] = stmt.Right.Kind.Value()
	}
	return assignments
}

func getValueType(localScope *scope.LocalScope, stmt *parser.DeclStmt) definition.BaseValueType {
	primitive, ok := definition.ParsePrimitive(stmt.Token.Literal)
	if ok {
		primitiveValueType := definition.PrimitiveValueType{
			Primitive: primitive,
			Nullable:  stmt.Nullable,
		}

		if len(stmt.Args) > 0 {
			primitiveValueType.Arguments = make([]definition.BaseValueType, len(stmt.Args))
			for i, arg := range stmt.Args {
				primitiveValueType.Arguments[i] = getValueType(localScope, &arg)
			}
		}

		return primitiveValueType
	}

	name, alias := stmt.Format()
	obj := localScope.MustFindObject(name, alias)
	return definition.CustomValueType{
		Nullable: stmt.Nullable,
		ObjectId: obj.Id,
	}
}
