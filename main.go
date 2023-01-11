package main

import (
	"encoding/json"

	"tomasweigenast.com/schema_interpreter/cmd"
)

func main() {
	builder := cmd.NewBuilder()
	err := builder.Build("testdata")
	if err != nil {
		println(err.Error())
	} else {
		// print definition
		def := builder.GetBuiltDefinition()
		println("definition hashcode: ", def.Hashcode)
		println("======================")

		buf, _ := json.MarshalIndent(def, "", "    ")
		println(string(buf))
	}
}
