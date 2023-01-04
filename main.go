package main

import (
	"bufio"
	"os"

	"tomasweigenast.com/schema_interpreter/internal"
)

func main() {
	f, _ := os.OpenFile("test", os.O_RDONLY, os.ModePerm)
	internal.NewTokenizer(bufio.NewReader(f))
}
