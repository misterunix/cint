package main

import (
	"fmt"

	"github.com/misterunix/cint"
)

func main() {
	source := `
int factorial(int n) {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}

int main() {
	printf("Starting...\n");
	int result = factorial(5);
	printf("Factorial of 5 is: %d\n", result);
	return 0;
}
`

	interp, err := cint.New(source)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	err = interp.Run()
	if err != nil {
		fmt.Println("Runtime error:", err)
		return
	}

	fmt.Println("Success!")
}
