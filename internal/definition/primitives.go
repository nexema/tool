package definition

type ValuePrimitive string

const (
	Undefined ValuePrimitive = ""
	String    ValuePrimitive = "string"
	Boolean   ValuePrimitive = "bool"
	Int       ValuePrimitive = "varint"
	Uint      ValuePrimitive = "uvarint"
	Int8      ValuePrimitive = "int8"
	Int16     ValuePrimitive = "int16"
	Int32     ValuePrimitive = "int32"
	Int64     ValuePrimitive = "int64"
	Uint8     ValuePrimitive = "uint8"
	Uint16    ValuePrimitive = "uint16"
	Uint32    ValuePrimitive = "uint32"
	Uint64    ValuePrimitive = "uint64"
	Float32   ValuePrimitive = "float32"
	Float64   ValuePrimitive = "float64"
	Map       ValuePrimitive = "map"
	List      ValuePrimitive = "list"
	Custom    ValuePrimitive = "custom" // User defined types
	Timestamp ValuePrimitive = "timestamp"
	Duration  ValuePrimitive = "duration"
)

var mapping map[string]ValuePrimitive = map[string]ValuePrimitive{
	"string":    String,
	"bool":      Boolean,
	"varint":    Int,
	"uvarint":   Uint,
	"int":       Int,
	"uint":      Uint,
	"int8":      Int8,
	"int16":     Int16,
	"int32":     Int32,
	"int64":     Int64,
	"uint8":     Uint8,
	"uint16":    Uint16,
	"uint32":    Uint32,
	"uint64":    Uint64,
	"float32":   Float32,
	"float64":   Float64,
	"map":       Map,
	"list":      List,
	"timestamp": Timestamp,
	"duration":  Duration,
}

func ParsePrimitive(name string) (ValuePrimitive, bool) {
	value, ok := mapping[name]
	return value, ok
}
