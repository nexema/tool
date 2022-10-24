package internal

import (
	"crypto/md5"
	"encoding/hex"

	b64 "encoding/base64"
)

func Skip[T comparable](arr []T, until T) []T {
	for i := 0; i < len(arr); i++ {
		if arr[i] == until {
			return arr[i:]
		}

	}

	return arr
}

func GetHash(typeName string, pkg string) string {
	typeNameBuf := []byte(typeName)
	pkgBuf := []byte(pkg)

	hash := md5.Sum(append(typeNameBuf, pkgBuf...))
	return hex.EncodeToString(hash[:])
}

func EncodeBase64(input string) string {
	return b64.StdEncoding.EncodeToString([]byte(input))
}

func GetAsPointer[T any](value T) *T {
	return &value
}

type TreePrintBoxType int

const (
	Regular TreePrintBoxType = iota
	Last
	AfterLast
	Between
)

func (boxType TreePrintBoxType) String() string {
	switch boxType {
	case Regular:
		return "\u251c" // ├
	case Last:
		return "\u2514" // └
	case AfterLast:
		return " "
	case Between:
		return "\u2502" // │
	default:
		panic("invalid box type")
	}
}

func getTreePrintBoxType(index int, len int) TreePrintBoxType {
	if index+1 == len {
		return Last
	} else if index+1 > len {
		return AfterLast
	}
	return Regular
}

func getTreePrintBoxTypeExternal(index int, len int) TreePrintBoxType {
	if index+1 == len {
		return AfterLast
	}
	return Between
}

func getTreePrintPadding(root bool, boxType TreePrintBoxType) string {
	if root {
		return ""
	}

	return boxType.String() + " "
}
