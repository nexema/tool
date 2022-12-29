package main

import (
	"bytes"
	"fmt"

	"tomasweigenast.com/schema_interpreter/internal"
)

func main() {
	input := `
	@metadata
	type MyName struct {}`
	parser := internal.NewParser(bytes.NewBufferString(input))
	ast, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
	}

	_ = ast
}
