package main

import (
	"fmt"

	"github.com/bjones/cint"
)

func main() {
	// Simplest possible C program
	source := `
int main() {
	return 0;
}
`

	fmt.Println("Source code:")
	fmt.Println(source)
	fmt.Println()

	interp, err := cint.New(source)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	fmt.Println("Parsed successfully!")

	if err := interp.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	} else {
		fmt.Println("Executed successfully!")
	}
}
