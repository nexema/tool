package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ValueParser struct{}

func NewValueParser() *ValueParser {
	return &ValueParser{}
}

// ParseString parses a string returning a Go value
func (v *ValueParser) ParseString(s string) (value interface{}, valueType FieldTypePrimitive, err error) {
	return v.internalParse(s, 0)
}

// CheckType checks if the given string contains a valid Go type.
// If parsed type does not match t or an error happens during the parse, an error is returned.
func (v *ValueParser) CheckType(s string, t FieldTypePrimitive) error {
	_, parsedType, err := v.internalParse(s, 0)
	if err != nil {
		return err
	}

	if parsedType == t {
		return nil
	} else {
		return fmt.Errorf("string %s expected to be type of %v but was %v", s, t, parsedType)
	}
}

// / internalParse parses a string into a FieldTypePrimitive
func (v *ValueParser) internalParse(s string, deep int) (value interface{}, valueType FieldTypePrimitive, err error) {
	s = strings.TrimSpace(s)

	firstChar := s[0]

	// Maybe a string
	if firstChar == '"' {
		endChar := s[len(s)-1]
		if endChar != '"' {
			return nil, UnknownFieldType, errors.New("an string must start and end with a \"")
		}

		return s[1 : len(s)-1], String, nil
	}

	// Maybe a boolean
	if s == "true" {
		return true, Boolean, nil
	}

	if s == "false" {
		return false, Boolean, nil
	}

	// Maybe an int
	number, err := strconv.Atoi(s)
	if err == nil {
		return number, Int64, nil
	}

	// Maybe a float
	float, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return float, Float64, nil
	}

	// Maybe a list
	listOrMap, valueType, ok, err := v.parseListOrMap(s, deep)
	if err != nil {
		return nil, UnknownFieldType, err
	}

	if ok {
		return listOrMap, valueType, nil
	}

	// Maybe a map

	return nil, UnknownFieldType, fmt.Errorf("cannot parse string to Go type, given: %v", s)
}

// parseListOrMap parses a string into a list or map
func (v *ValueParser) parseListOrMap(s string, deep int) (mapOrList interface{}, t FieldTypePrimitive, ok bool, err error) {
	firstChar := s[0]
	lastChar := s[len(s)-1]

	// its a list or a map
	if firstChar == '[' && lastChar == ']' {
		if deep != 0 {
			return nil, UnknownFieldType, false, errors.New("cannot nest list or maps")
		}

		// due to a map is only a list of key-value pairs, they have the same signature, both starts with double braces: "[]"
		tokens := strings.Split(s[1:len(s)-1], ",") // get all tokens

		// empty list or map as default values are not allowed
		if len(tokens) == 0 {
			return nil, UnknownFieldType, false, errors.New("expected a list or map to have at least one item")
		}

		var listValue []interface{}
		var mapValue map[interface{}]interface{}

		// TODO: the check to determine if a map or list can be improved using regex
		for i, token := range tokens {

			token = strings.TrimSpace(token)

			// If a token contains a ":", its a map
			if strings.Contains(token, ":") {
				if mapValue == nil {
					mapValue = make(map[interface{}]interface{})
				}

				key, _, value, _, err := v.parseMapEntry(token[1 : len(token)-1])
				if err != nil {
					return nil, UnknownFieldType, false, err
				}

				mapValue[key] = value

			} else { // its a list
				if listValue == nil {
					listValue = make([]interface{}, len(tokens))
				}

				// Parse token
				value, _, err := v.internalParse(token, 1)
				if err != nil {
					return nil, UnknownFieldType, false, fmt.Errorf("cannot parse list item: %s", err.Error())
				}

				listValue[i] = value
			}
		}

		if listValue == nil {
			return mapValue, Map, true, nil
		} else {
			return listValue, List, true, nil
		}

	} else {
		return nil, UnknownFieldType, false, nil
	}
}

// parseMapEntry parses s into a map entry
func (v *ValueParser) parseMapEntry(s string) (key interface{}, keyType FieldTypePrimitive, value interface{}, valueType FieldTypePrimitive, err error) {
	tokens := strings.Split(s, ":")
	if len(tokens) != 2 {
		return nil, UnknownFieldType, nil, UnknownFieldType, fmt.Errorf("map entries must have two values, a key and a value, given: %v", s)
	}

	rawKey := tokens[0]
	rawValue := tokens[1]

	// parse key
	key, keyType, err = v.internalParse(rawKey, 0)
	if err != nil {
		return nil, UnknownFieldType, nil, UnknownFieldType, err
	}

	value, valueType, err = v.internalParse(rawValue, 0)
	if err != nil {
		return nil, UnknownFieldType, nil, UnknownFieldType, err
	}

	return key, keyType, value, valueType, nil
}
