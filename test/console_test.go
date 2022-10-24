package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tomasweigenast.com/schema_interpreter/internal"
)

func TestConsoleBuild(t *testing.T) {
	t.Skip()
	err := internal.ConsoleBuild(internal.NewBuilder(), "json", "console", "./test_files")
	assert.NoError(t, err)
}

func TestConsoleGenerate(t *testing.T) {
	t.Skip()
	err := internal.ConsoleGenerate(internal.NewBuilder(), "./test_files", "./src/models/dto", "json")
	assert.NoError(t, err)
}
