# C Interpreter - Quick Reference

## Installation & Import
```go
import "github.com/bjones/cint"
```

## Basic Usage

### Parse and Run
```go
source := `int main() { return 0; }`
interp, err := cint.New(source)
if err != nil {
    // Handle parse error
}
err = interp.Run()
if err != nil {
    // Handle runtime error
}
```

### Single-Stepping
```go
interp.EnableSingleStep()
for {
    result := interp.Step()
    if result.Error != nil {
        // Handle error
    }
    if result.Done {
        break
    }
    // Process step...
}
```

### Reset
```go
interp.Reset()  // Reset to initial state
```

## Built-in Functions

### printf
```c
printf("Hello %s, value: %d\n", "world", 42);
```

### sleep (milliseconds)
```c
sleep(1000);  // Sleep for 1 second
```

### putchar
```c
putchar('A');  // Output: A
```

## Supported C Features

### Data Types
- int, float, double, char, void

### Operators
- Arithmetic: +, -, *, /, %
- Comparison: ==, !=, <, >, <=, >=
- Logical: &&, ||, !
- Bitwise: &, |, ^, ~, <<, >>
- Assignment: =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=
- Inc/Dec: ++, --
- Ternary: ? :

### Control Flow
- if / else
- while loops
- for loops
- break, continue
- return

### Functions
- Function declarations
- Function calls
- Recursion

## Example Programs

### Simple
```c
int main() {
    int x = 10;
    printf("x = %d\n", x);
    return 0;
}
```

### With Sleep
```c
int main() {
    int i;
    for (i = 0; i < 5; i++) {
        printf("Count: %d\n", i);
        sleep(500);
    }
    return 0;
}
```

### Recursive
```c
int factorial(int n) {
    if (n <= 1) return 1;
    return n * factorial(n - 1);
}

int main() {
    printf("Result: %d\n", factorial(5));
    return 0;
}
```

## API Reference

### cint.New(source string) (*Cint, error)
Creates a new interpreter from C source code.

### (*Cint).Run() error
Executes the program starting from main().

### (*Cint).EnableSingleStep()
Enables single-step execution mode.

### (*Cint).DisableSingleStep()
Disables single-step execution mode.

### (*Cint).Step() *StepResult
Executes one statement in single-step mode.

### (*Cint).Reset()
Resets the interpreter to initial state.

### StepResult
```go
type StepResult struct {
    Statement Statement
    Line      int
    Done      bool      // Execution complete?
    Returned  bool      // Return executed?
    ReturnVal *Value    // Return value
    Break     bool      // Break executed?
    Continue  bool      // Continue executed?
    Error     error     // Any error
}
```

## Common Patterns

### Error Handling
```go
interp, err := cint.New(source)
if err != nil {
    fmt.Printf("Parse error: %v\n", err)
    return
}

if err := interp.Run(); err != nil {
    fmt.Printf("Runtime error: %v\n", err)
    return
}
```

### Debugging
```go
interp.EnableSingleStep()
stepNum := 1
for {
    result := interp.Step()
    if result.Error != nil {
        fmt.Printf("Error at step %d: %v\n", stepNum, result.Error)
        break
    }
    if result.Done {
        fmt.Println("Program completed")
        break
    }
    fmt.Printf("Step %d executed\n", stepNum)
    stepNum++
}
```

## Files

- `cint.go` - Main API
- `lexer.go` - Tokenizer
- `parser.go` - Parser
- `interpreter.go` - Interpreter
- `ast.go` - AST definitions
- `token.go` - Token definitions

## Examples Location
`/home/bjones/go/src/cint/examples/`
- basic/
- debugger/
- module_test/
- test_suite/

## Build & Test
```bash
cd /home/bjones/go/src/cint
go build ./...
go run examples/module_test/main.go
```
