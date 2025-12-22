package main

import (
	"fmt"

	"github.com/bjones/cint"
)

func main() {
	// Example 1: Basic math functions
	fmt.Println("=== Example 1: Basic Math Functions ===")
	source1 := `
	int main() {
		printf("sqrt(16.0) = %f\n", sqrt(16.0));
		printf("sqrt(2.0) = %f\n", sqrt(2.0));
		printf("pow(2.0, 3.0) = %f\n", pow(2.0, 3.0));
		printf("pow(10.0, 2.0) = %f\n", pow(10.0, 2.0));
		printf("abs(-5) = %d\n", abs(-5));
		printf("abs(-3.5) = %f\n", abs(-3.5));
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

	// Example 2: Trigonometric functions
	fmt.Println("\n=== Example 2: Trigonometric Functions ===")
	source2 := `
	int main() {
		float pi = 3.14159265;
		printf("sin(0) = %f\n", sin(0.0));
		printf("sin(pi/2) = %f\n", sin(pi / 2.0));
		printf("cos(0) = %f\n", cos(0.0));
		printf("cos(pi) = %f\n", cos(pi));
		printf("tan(0) = %f\n", tan(0.0));
		printf("tan(pi/4) = %f\n", tan(pi / 4.0));
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

	// Example 3: Rounding functions
	fmt.Println("\n=== Example 3: Rounding Functions ===")
	source3 := `
	int main() {
		float num = 3.7;
		printf("floor(3.7) = %f\n", floor(3.7));
		printf("floor(-3.7) = %f\n", floor(-3.7));
		printf("ceil(3.2) = %f\n", ceil(3.2));
		printf("ceil(-3.2) = %f\n", ceil(-3.2));
		return 0;
	}
	`

	interp3, err := cint.New(source3)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	if err := interp3.Run(); err != nil {
		fmt.Println("Runtime error:", err)
	}

	// Example 4: Logarithmic and exponential functions
	fmt.Println("\n=== Example 4: Logarithmic and Exponential Functions ===")
	source4 := `
	int main() {
		printf("exp(1.0) = %f\n", exp(1.0));
		printf("exp(2.0) = %f\n", exp(2.0));
		printf("log(2.71828) = %f\n", log(2.71828));
		printf("log(10.0) = %f\n", log(10.0));
		printf("log10(100.0) = %f\n", log10(100.0));
		printf("log10(1000.0) = %f\n", log10(1000.0));
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

	// Example 5: Practical example - calculating distance and area
	fmt.Println("\n=== Example 5: Practical Example - Distance and Area ===")
	source5 := `
	int main() {
		float x1 = 0.0;
		float y1 = 0.0;
		float x2 = 3.0;
		float y2 = 4.0;
		
		float dx = x2 - x1;
		float dy = y2 - y1;
		float distance = sqrt(dx * dx + dy * dy);
		
		printf("Distance between (0,0) and (3,4): %f\n", distance);
		
		// Calculate area of circle with radius 5
		float radius = 5.0;
		float pi = 3.14159265;
		float area = pi * pow(radius, 2.0);
		
		printf("Area of circle with radius 5: %f\n", area);
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
