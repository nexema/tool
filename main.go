package main

import (
	"tomasweigenast.com/nexema/tool/cmd"
)

func main() {
	cmd.Run()
	// builder := cmd.NewBuilder()
	// err := builder.Build("testdata")
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// } /*else {
	// 	// print definition
	// 	def := builder.GetBuiltDefinition()
	// 	println("definition hashcode: ", def.Hashcode)
	// 	println("======================")

	// 	buf, _ := json.MarshalIndent(def, "", "    ")
	// 	println(string(buf))
	// } */

	// buf, err := json.Marshal(builder.GetBuiltDefinition())
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }

	// plugin := cmd.NewPlugin("go", `/Users/tomasweigenast/Desktop/Git/nexema/go/go`)
	// err = plugin.Run(buf)
	// if err != nil {
	// 	println(err.Error())
	// } else {
	// 	println("success")
	// }
}
