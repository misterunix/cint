# C Interpreter (cint)

A C interpreter written in Go that follows the K&R (Kernighan & Ritchie) C standard. This interpreter can be used as a Go module to execute C code dynamically with support for single-stepping execution.

## Notes

All K&R test appear to work. If you find a problem, contribute, please!

## Features

- **K&R C Standard Compliance**: Implements core features of the original C language
- **Single-Step Execution**: Step through code line by line for debugging
- **Built-in Functions**: Including `printf`, `sleep` (with millisecond resolution), and `putchar`
- **Module Interface**: Designed to be imported and used by other Go programs
- **Full Expression Support**: Arithmetic, logical, bitwise, and comparison operators
- **Control Flow**: if/else, while, for loops, break, continue, return
- **Function Declarations**: Support for user-defined functions with parameters

## Installation

```bash
go get github.com/misterunix/cint
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/misterunix/cint"
)

func main() {
    source := `
    int main() {
        int x = 10;
        int y = 20;
        printf("Sum: %d\n", x + y);
        return 0;
    }
    `
    
    interp, err := cint.New(source)
    if err != nil {
        fmt.Println("Parse error:", err)
        return
    }
    
    if err := interp.Run(); err != nil {
        fmt.Println("Runtime error:", err)
    }
}
```

### Single-Step Execution

```go
interp, _ := cint.New(source)
interp.EnableSingleStep()

for {
    result := interp.Step()
    if result.Done {
        break
    }
    if result.Error != nil {
        fmt.Println("Error:", result.Error)
        break
    }
}
```

### Using the Sleep Function

The interpreter includes a `sleep()` function with millisecond resolution:

```c
int main() {
    int i;
    for (i = 0; i < 5; i++) {
        printf("Count: %d\n", i);
        sleep(1000);  // Sleep for 1000 milliseconds (1 second)
    }
    return 0;
}
```

## API Reference

### Creating an Interpreter

```go
func New(source string) (*Cint, error)
```

Creates a new interpreter instance from C source code.

### Running Code

```go
func (c *Cint) Run() error
```

Executes the entire program starting from `main()`.

### Single-Stepping

```go
func (c *Cint) EnableSingleStep()
func (c *Cint) DisableSingleStep()
func (c *Cint) Step() *StepResult
func (c *Cint) Reset()
```

- `EnableSingleStep()`: Enables single-step mode
- `DisableSingleStep()`: Disables single-step mode
- `Step()`: Executes one statement and returns the result
- `Reset()`: Resets the interpreter state

### StepResult Structure

```go
type StepResult struct {
    Statement Statement  // The statement that was executed
    Line      int       // Line number
    Done      bool      // Whether execution is complete
    Returned  bool      // Whether a return was executed
    ReturnVal *Value    // Return value if any
    Break     bool      // Whether break was executed
    Continue  bool      // Whether continue was executed
    Error     error     // Any error that occurred
}
```

## Built-in Functions

### printf

Standard formatted output:

```c
printf("Value: %d\n", 42);
printf("Float: %f\n", 3.14);
printf("String: %s\n", "hello");
```

### sleep

Sleep with millisecond resolution:

```c
sleep(500);  // Sleep for 500 milliseconds
```

### putchar

Output a single character:

```c
putchar('A');  // Outputs: A
```

## Supported C Features

### Data Types
- `int`
- `float`
- `double`
- `char`
- `void`

### Operators
- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`
- Bitwise: `&`, `|`, `^`, `~`, `<<`, `>>`
- Assignment: `=`, `+=`, `-=`, `*=`, `/=`, `%=`
- Increment/Decrement: `++`, `--`
- Ternary: `? :`

### Control Flow
- `if` / `else`
- `while` loops
- `for` loops
- `break` and `continue`
- `return`

### Functions
- Function declarations with parameters
- Function calls
- Recursion support

## Examples

### Running Examples

```bash
# Basic examples
cd examples/basic
go run main.go

# Interactive debugger
cd examples/debugger
go run main.go
```

### Factorial Example

```c
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
```

## Architecture

The interpreter consists of several components:

1. **Lexer** (`lexer.go`): Tokenizes C source code
2. **Parser** (`parser.go`): Builds an Abstract Syntax Tree (AST)
3. **AST** (`ast.go`): Defines the structure of C code
4. **Interpreter** (`interpreter.go`): Executes the AST with single-stepping support
5. **API** (`cint.go`): Public interface for using the interpreter as a module

## Limitations

This interpreter implements a subset of K&R C:

- No preprocessor directives (#include, #define, etc.)
- No structs or unions
- No pointers (partially implemented)
- No arrays (basic support only)
- Limited standard library functions
- No file I/O

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.
