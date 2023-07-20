package schema

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/hashstructure/v2"
	"tomasweigenast.com/nexema/tool/internal/definition"
	"tomasweigenast.com/nexema/tool/internal/parser"
	"tomasweigenast.com/nexema/tool/internal/scope"
	"tomasweigenast.com/nexema/tool/internal/token"
)

const (
	SCHEMA_VERSION_1 int = 1
)

// SchemaBuilder converts a already analyzed ast to a valid Nexema definition.
type SchemaBuilder struct {
	root scope.Scope
}

func NewSchemaBuilder(rootScope scope.Scope) *SchemaBuilder {
	return &SchemaBuilder{rootScope}
}

// BuildSnapshot creates a Nexema Snapshot
func (self *SchemaBuilder) BuildSnapshot() *definition.NexemaSnapshot {
	snapshot := &definition.NexemaSnapshot{
		Version: SCHEMA_VERSION_1,
		Files:   make([]definition.NexemaFile, 0),
	}

	ids := make([]string, 0)

	self.buildPackage(self.root.(*scope.PackageScope), snapshot, &ids)
	snapshotHashcode, err := hashstructure.Hash(&ids, hashstructure.FormatV2, &hashstructure.HashOptions{})
	if err != nil {
		panic(fmt.Errorf("could not calculate hash for snapshot. Please report this because this is not expected. Error: %v", err))
	}

	snapshot.Hashcode = fmt.Sprint(snapshotHashcode)

	return snapshot
}

func (self *SchemaBuilder) buildPackage(s *scope.PackageScope, snapshot *definition.NexemaSnapshot, ids *[]string) {
	for _, child := range s.Children {
		if child.Kind() == scope.Package {
			self.buildPackage(child.(*scope.PackageScope), snapshot, ids)
		} else {
			file := self.buildFile(child.(*scope.FileScope))
			snapshot.Files = append(snapshot.Files, *file)
			*ids = append(*ids, file.Id)
		}
	}
}

func (self *SchemaBuilder) buildFile(fs *scope.FileScope) *definition.NexemaFile {
	physicalFile := fs.Path()
	file := &definition.NexemaFile{
		Path:        physicalFile,
		PackageName: fs.Parent().Name(),
		Types:       make([]definition.TypeDefinition, len(fs.Objects)),
	}

	idx := 0
	for _, obj := range fs.Objects {
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
			baseType := fs.FindObject(name, alias)
			if baseType == nil || len(baseType) != 1 {
				panic(fmt.Errorf("this should not happen, base type %s.%s not found", name, alias))
			}

			typeDef.BaseType = &baseType[0].Id
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
			}

			if stmt.Modifier != token.Enum {
				typeDef.Fields[i].Type = getValueType(fs, field.ValueType)
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

func getValueType(localScope *scope.FileScope, stmt *parser.DeclStmt) definition.BaseValueType {
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
	objs := localScope.FindObject(name, alias)
	if objs == nil || len(objs) != 1 {
		panic(fmt.Errorf("this should not happen, unable to find object %s.%s", alias, name))
	}
	return definition.CustomValueType{
		Nullable: stmt.Nullable,
		ObjectId: objs[0].Id,
	}
}
