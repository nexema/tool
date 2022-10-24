package internal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Tokenizer interface {
	Tokenize(line string) (tokens []string)
}

type StructFieldTokenizer struct{}

func NewStructFieldTokenizer() *StructFieldTokenizer {
	return &StructFieldTokenizer{}
}

type StructFieldTokenizerResult struct {
	FieldName              string
	PrimitiveFieldTypeName string
	TypeArguments          []string
	FieldIndex             string
	DefaultValue           string
	Metadata               string
}

var fieldNameTypeIndexMatcher = regexp.MustCompile(`([a-zA-Z][a-zA-Z\d]*):([a-zA-Z\d._]+\??)(\(([a-zA-Z\d._]+\??),? ?([a-zA-Z\d._]+\??)?\))? (\d+)`)
var defaultValueMatcher = regexp.MustCompile(`= ?(".*?"|[a-zA-Z0-9.]+|\[.[^@]+])`) // this should avoid special chars
var metadataMatcherRegex = regexp.MustCompile(`@(\[.+])`)                          // this should avoid special chars

// Tokenize tokenizes a string line searching for struct fields
func (s *StructFieldTokenizer) Tokenize(line string) (result *StructFieldTokenizerResult, err error) {
	line = strings.Join(strings.Fields(line), " ")

	result = &StructFieldTokenizerResult{}

	// match field name, type and index
	globalMatches := fieldNameTypeIndexMatcher.FindAllStringSubmatch(line, -1)

	if len(globalMatches) == 0 {
		return nil, fmt.Errorf("unexpected struct field syntax, given: %v", line)
	}

	matches := globalMatches[0]
	if len(matches) < 3 {
		return nil, fmt.Errorf("expected at least 3 arguments while declaring a field, given: %v", line)
	}

	result.FieldName = matches[1]
	result.PrimitiveFieldTypeName = matches[2]
	result.FieldIndex = matches[6]

	firstTypeArgument := matches[4]
	secondTypeArgument := matches[5]

	if len(firstTypeArgument) > 0 {
		if result.TypeArguments == nil {
			result.TypeArguments = make([]string, 0)
		}

		result.TypeArguments = append(result.TypeArguments, firstTypeArgument)
	}

	if len(secondTypeArgument) > 0 {
		if result.TypeArguments == nil {
			result.TypeArguments = make([]string, 0)
		}

		result.TypeArguments = append(result.TypeArguments, secondTypeArgument)
	}

	// remove from string
	index := fieldNameTypeIndexMatcher.FindAllStringIndex(line, -1)[0][1]
	line = strings.TrimSpace(line[index:])
	// match default value
	defaultValueMatch := defaultValueMatcher.FindAllStringSubmatch(line, 1)

	if len(defaultValueMatch) == 1 {
		defaultValue := defaultValueMatch[0]

		if len(defaultValue) == 2 {
			result.DefaultValue = defaultValue[1]
			index = defaultValueMatcher.FindAllStringIndex(line, -1)[0][1]
			line = line[index:]
		} else if len(defaultValue) > 2 {
			return nil, errors.New("only one default value is allowed")
		}
	}

	// match metadata
	metadataMatcher := metadataMatcherRegex.FindAllStringSubmatch(line, -1)
	if len(metadataMatcher) == 1 {
		metadata := metadataMatcher[0]
		if len(metadata) == 2 {
			result.Metadata = metadata[1]
		} else if len(metadata) > 2 {
			return nil, errors.New("only one metadata value is allowed")
		}
	}

	return result, nil
}
