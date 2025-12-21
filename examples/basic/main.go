package main

import (
	"fmt"

	"github.com/bjones/cint"
)

func main() {
	// Example 1: Simple program
	fmt.Println("=== Example 1: Simple Program ===")
	source1 := `
	int main() {
		int x = 10;
		int y = 20;
		int sum = x + y;
		printf("Sum: %d\n", sum);
		return 0;
	}
	`

	interp1, err := cint.New(source1)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if err := interp1.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}

	// Example 2: Loop with sleep
	fmt.Println("\n=== Example 2: Loop with Sleep ===")
	source2 := `
	int main() {
		int i;
		for (i = 0; i < 5; i++) {
			printf("Count: %d\n", i);
			sleep(500);
		}
		return 0;
	}
	`

	interp2, err := cint.New(source2)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if err := interp2.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}

	// Example 3: Single-stepping
	fmt.Println("\n=== Example 3: Single-Stepping ===")
	source3 := `
	int main() {
		int x = 5;
		int y = x * 2;
		printf("Result: %d\n", y);
		return 0;
	}
	`

	interp3, err := cint.New(source3)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	interp3.EnableSingleStep()

	stepNum := 1
	for {
		result := interp3.Step()
		if result.Error != nil {
			fmt.Printf("Error at step %d: %v\n", stepNum, result.Error)
			break
		}

		if result.Done {
			fmt.Println("Program completed")
			if result.Returned && result.ReturnVal != nil {
				fmt.Printf("Return value: %d\n", result.ReturnVal.Int)
			}
			break
		}

		fmt.Printf("Step %d: Executed statement\n", stepNum)
		stepNum++
	}

	// Example 4: Function calls
	fmt.Println("\n=== Example 4: Function Calls ===")
	source4 := `
	int factorial(int n) {
		if (n <= 1) {
			return 1;
		}
		return n * factorial(n - 1);
	}
	
	int main() {
		int result = factorial(5);
		printf("Factorial of 5 is: %d\n", result);
		return 0;
	}
	`

	interp4, err := cint.New(source4)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if err := interp4.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}

	// Example 5: Conditional and operators
	fmt.Println("\n=== Example 5: Conditionals and Operators ===")
	source5 := `
	int main() {
		int a = 10;
		int b = 20;
		
		if (a < b) {
			printf("a is less than b\n");
		} else {
			printf("a is greater or equal to b\n");
		}
		
		int max = (a > b) ? a : b;
		printf("Max value: %d\n", max);
		
		return 0;
	}
	`

	interp5, err := cint.New(source5)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if err := interp5.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}
}
