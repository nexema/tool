package definition

type FieldDefinition struct {
	Name          string
	Index         int
	Type          BaseValueType
	Documentation []string
	Annotations   Assignments
}

type BaseValueTypeKind string

const (
	PrimitiveKind BaseValueTypeKind = "primitiveValueType"
	CustomKind    BaseValueTypeKind = "customType"
)

type BaseValueType interface {
	Kind() BaseValueTypeKind
}

type PrimitiveValueType struct {
	Primitive ValuePrimitive
	Nullable  bool
	Arguments []BaseValueType
}

type CustomValueType struct {
	ObjectId uint64
	Nullable bool
}

func (PrimitiveValueType) Kind() BaseValueTypeKind {
	return PrimitiveKind
}

func (CustomValueType) Kind() BaseValueTypeKind {
	return CustomKind
}
