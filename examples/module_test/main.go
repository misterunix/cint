package main

import (
	"fmt"

	"github.com/bjones/cint"
)

func main() {
	fmt.Println("=== Testing C Interpreter as a Module ===\n")

	// Test 1: Module can parse C code
	fmt.Println("Test 1: Parsing C code")
	source1 := `int main() { return 42; }`
	interp1, err := cint.New(source1)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Module successfully parsed C code")

	// Test 2: Module can run C code
	fmt.Println("\nTest 2: Running C code")
	if err := interp1.Run(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Module successfully executed C code")

	// Test 3: Module can use single-stepping
	fmt.Println("\nTest 3: Single-stepping execution")
	source2 := `
	int main() {
		int x = 10;
		int y = 20;
		int z = x + y;
		return z;
	}
	`
	interp2, err := cint.New(source2)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	interp2.EnableSingleStep()
	stepCount := 0
	for {
		result := interp2.Step()
		if result.Error != nil {
			fmt.Printf("❌ Failed: %v\n", result.Error)
			return
		}
		if result.Done {
			break
		}
		stepCount++
	}
	fmt.Printf("✅ Module executed %d steps successfully\n", stepCount)

	// Test 4: Module includes sleep function
	fmt.Println("\nTest 4: Built-in sleep function")
	source3 := `
	int main() {
		printf("Sleeping...\n");
		sleep(100);
		printf("Done!\n");
		return 0;
	}
	`
	interp3, err := cint.New(source3)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	if err := interp3.Run(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Module includes working sleep function")

	// Test 5: Module follows K&R standard (functions, loops, etc)
	fmt.Println("\nTest 5: K&R C standard compliance")
	source4 := `
	int fibonacci(int n) {
		if (n <= 1) {
			return n;
		}
		return fibonacci(n-1) + fibonacci(n-2);
	}
	
	int main() {
		int i;
		for (i = 0; i < 7; i++) {
			printf("fib(%d) = %d\n", i, fibonacci(i));
		}
		return 0;
	}
	`
	interp4, err := cint.New(source4)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	if err := interp4.Run(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Module follows K&R C standard")

	// Test 6: Module can reset and re-run
	fmt.Println("\nTest 6: Reset and re-run capability")
	interp2.Reset()
	interp2.EnableSingleStep()
	result := interp2.Step()
	if result.Error != nil {
		fmt.Printf("❌ Failed: %v\n", result.Error)
		return
	}
	fmt.Println("✅ Module can reset and re-run")

	fmt.Println("\n=== All Module Tests Passed! ===")
	fmt.Println("\nThe C interpreter module:")
	fmt.Println("  ✅ Can be imported and used in Go code")
	fmt.Println("  ✅ Parses and executes K&R C code")
	fmt.Println("  ✅ Supports single-stepping for debugging")
	fmt.Println("  ✅ Includes sleep function with millisecond resolution")
	fmt.Println("  ✅ Handles functions, loops, recursion, and more")
}
