package main

import (
	"fmt"

	"github.com/misterunix/cint"
)

func main() {
	source := `
int main() {
	int i;
	for (i = 0; i < 5; i++) {
		printf("i=%d\n", i);
	}
	return 0;
}
`

	fmt.Println("Testing variable declaration...")
	interp, err := cint.New(source)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	fmt.Println("Parse successful!")

	if err := interp.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}
}
