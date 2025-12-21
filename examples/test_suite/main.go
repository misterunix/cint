package main

import (
	"fmt"

	"github.com/bjones/cint"
)

func testBasic() bool {
	source := `
	int main() {
		int x = 10;
		int y = 20;
		return x + y;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Basic test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Basic test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Basic test passed")
	return true
}

func testSleep() bool {
	source := `
	int main() {
		printf("Starting sleep test\n");
		sleep(100);
		printf("After 100ms sleep\n");
		return 0;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Sleep test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Sleep test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Sleep test passed")
	return true
}

func testSingleStepping() bool {
	source := `
	int main() {
		int x = 5;
		int y = x * 2;
		return y;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Single-stepping test failed: %v\n", err)
		return false
	}

	interp.EnableSingleStep()
	steps := 0
	for {
		result := interp.Step()
		if result.Error != nil {
			fmt.Printf("❌ Single-stepping test failed: %v\n", result.Error)
			return false
		}
		if result.Done {
			break
		}
		steps++
	}

	if steps < 2 {
		fmt.Printf("❌ Single-stepping test failed: expected at least 2 steps, got %d\n", steps)
		return false
	}

	fmt.Printf("✅ Single-stepping test passed (%d steps)\n", steps)
	return true
}

func testRecursion() bool {
	source := `
	int factorial(int n) {
		if (n <= 1) {
			return 1;
		}
		return n * factorial(n - 1);
	}
	
	int main() {
		int result = factorial(6);
		printf("Factorial(6) = %d\n", result);
		return 0;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Recursion test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Recursion test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Recursion test passed")
	return true
}

func testLoops() bool {
	source := `
	int main() {
		int sum = 0;
		int i;
		for (i = 1; i <= 5; i++) {
			sum = sum + i;
		}
		printf("Sum 1-5 = %d\n", sum);
		return sum;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Loops test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Loops test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Loops test passed")
	return true
}

func testConditionals() bool {
	source := `
	int main() {
		int x = 10;
		int y = 20;
		int max;
		
		if (x > y) {
			max = x;
		} else {
			max = y;
		}
		
		printf("Max of %d and %d is %d\n", x, y, max);
		return max;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Conditionals test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Conditionals test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Conditionals test passed")
	return true
}

func testOperators() bool {
	source := `
	int main() {
		int a = 10;
		int b = 3;
		int c;
		
		c = a + b;
		c = a - b;
		c = a * b;
		c = a / b;
		c = a % b;
		c = a & b;
		c = a | b;
		c = a ^ b;
		c = a << 1;
		c = a >> 1;
		
		int eq = (a == b);
		int ne = (a != b);
		int lt = (a < b);
		int gt = (a > b);
		
		a++;
		b--;
		
		printf("Operators test completed\n");
		return 0;
	}
	`
	interp, err := cint.New(source)
	if err != nil {
		fmt.Printf("❌ Operators test failed: %v\n", err)
		return false
	}
	if err := interp.Run(); err != nil {
		fmt.Printf("❌ Operators test failed: %v\n", err)
		return false
	}
	fmt.Println("✅ Operators test passed")
	return true
}

func main() {
	fmt.Println("=== C Interpreter Test Suite ===\n")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Basic Execution", testBasic},
		{"Sleep Function", testSleep},
		{"Single-Stepping", testSingleStepping},
		{"Recursion", testRecursion},
		{"Loops", testLoops},
		{"Conditionals", testConditionals},
		{"Operators", testOperators},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		fmt.Printf("Running: %s\n", test.name)
		if test.fn() {
			passed++
		} else {
			failed++
		}
		fmt.Println()
	}

	fmt.Println("=== Test Results ===")
	fmt.Printf("Passed: %d/%d\n", passed, len(tests))
	fmt.Printf("Failed: %d/%d\n", failed, len(tests))

	if failed == 0 {
		fmt.Println("\n🎉 All tests passed!")
	}
}
