package main

import (
	"bufio"
	"bytes"
	"fmt"
)

func main() {
	reader := bufio.NewReader(bytes.NewBufferString("¢∞¬÷"))
	ch, size, err := reader.ReadRune()
	fmt.Printf("Character: %c\n", ch)
	fmt.Printf("Size: %d\n", size)
	fmt.Printf("Error: %s\n", err)
}
