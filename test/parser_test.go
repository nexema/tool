package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestParse(t *testing.T) {

	parser := internal.NewParser()
	tree, err := parser.ParseDirectory("test_files", "test_files")
	assert.Nil(t, err)

	if err == nil {
		t.Logf("\n%v", tree.Sprint())
	} else {
		return
	}

	typeResolver := internal.NewTypeResolver()
	err = typeResolver.Resolve(tree)
	assert.Nil(t, err)
}
