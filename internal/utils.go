package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

// hash hashes input using sha256
var hashInst hash.Hash = sha256.New()

func hashString(input string) string {
	hashInst.Reset()
	hashInst.Write([]byte(input))
	bs := hashInst.Sum(nil)
	return hex.EncodeToString(bs)
}

// intFits check if value fits "into" int primitive
func intFits(value int64, into Primitive) bool {
	switch into {
	case Primitive_Int64:
		return true

	case Primitive_Int32:
		return int64(int32(value)) == value

	case Primitive_Int16:
		return int64(int16(value)) == value

	case Primitive_Int8:
		return int64(int8(value)) == value

	case Primitive_Uint64:
		if value < 0 {
			return false
		}
		return int64(uint64(value)) == value

	case Primitive_Uint32:
		if value < 0 {
			return false
		}

		return int64(uint32(value)) == value

	case Primitive_Uint16:
		if value < 0 {
			return false
		}
		return int64(uint16(value)) == value

	case Primitive_Uint8:
		if value < 0 {
			return false
		}
		return int64(uint8(value)) == value

	case Primitive_Float32:
		return int64(float32(value)) == value

	case Primitive_Float64:
		return int64(float64(value)) == value

	default:
		panic("invalid int primitive")
	}
}

// floatFits check if value fits "into" float primitive
func floatFits(value float64, into Primitive) bool {
	switch into {
	case Primitive_Float64:
		return true

	case Primitive_Float32:
		return float64(float32(value)) == value

	default:
		panic("invalid float primitive")
	}
}
