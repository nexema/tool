package definition

type FieldDefinition struct {
	Name          string        `json:"name"`
	Index         int           `json:"index"`
	Type          BaseValueType `json:"type"`
	Documentation []string      `json:"documentation"`
	Annotations   Assignments   `json:"annotations"`
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
	Primitive ValuePrimitive  `json:"primitive"`
	Nullable  bool            `json:"nullable"`
	Arguments []BaseValueType `json:"arguments"`
}

type CustomValueType struct {
	ObjectId string `json:"objectId"`
	Nullable bool   `json:"nullable"`
}

func (PrimitiveValueType) Kind() BaseValueTypeKind {
	return PrimitiveKind
}

func (CustomValueType) Kind() BaseValueTypeKind {
	return CustomKind
}

func (self PrimitiveValueType) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"kind":      self.Kind(),
		"primitive": self.Primitive,
		"nullable":  self.Nullable,
		"arguments": self.Arguments,
	}
	return json.Marshal(m)
}

func (self CustomValueType) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"kind":     self.Kind(),
		"objectId": self.ObjectId,
		"nullable": self.Nullable,
	}
	return json.Marshal(m)
}
