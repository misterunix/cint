package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bjones/cint"
)

func main() {
	fmt.Println("Interactive C Interpreter Debugger")
	fmt.Println("Type C code and use 'step' to execute line by line")
	fmt.Println("Commands:")
	fmt.Println("  step - Execute one statement")
	fmt.Println("  run  - Run to completion")
	fmt.Println("  reset - Reset interpreter")
	fmt.Println("  quit - Exit")
	fmt.Println()

	// Example program with stepping capability
	source := `
	int main() {
		int sum = 0;
		int i;
		
		for (i = 1; i <= 10; i++) {
			sum = sum + i;
			printf("i=%d, sum=%d\n", i, sum);
		}
		
		printf("Final sum: %d\n", sum);
		return sum;
	}
	`

	fmt.Println("Loaded program:")
	fmt.Println(source)
	fmt.Println()

	interp, err := cint.New(source)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	interp.EnableSingleStep()

	scanner := bufio.NewScanner(os.Stdin)
	stepNum := 1

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())

		switch command {
		case "step", "s":
			result := interp.Step()

			if result.Error != nil {
				fmt.Printf("Error at step %d: %v\n", stepNum, result.Error)
				continue
			}

			if result.Done {
				fmt.Println("Program completed")
				if result.Returned && result.ReturnVal != nil {
					fmt.Printf("Return value: %d\n", result.ReturnVal.Int)
				}
				fmt.Println("Type 'reset' to restart or 'quit' to exit")
				continue
			}

			fmt.Printf("Step %d executed\n", stepNum)
			stepNum++

		case "run", "r":
			for {
				result := interp.Step()

				if result.Error != nil {
					fmt.Printf("Error: %v\n", result.Error)
					break
				}

				if result.Done {
					fmt.Println("Program completed")
					if result.Returned && result.ReturnVal != nil {
						fmt.Printf("Return value: %d\n", result.ReturnVal.Int)
					}
					break
				}
				stepNum++
			}

		case "reset":
			interp.Reset()
			stepNum = 1
			fmt.Println("Interpreter reset")

		case "quit", "q", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Available: step, run, reset, quit")
		}
	}
}
