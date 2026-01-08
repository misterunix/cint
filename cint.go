package cint

// Cint represents a wrapper around an Interpreter instance, providing methods and state
// for interacting with the interpreter in the context of the cint package.
type Cint struct {
	interpreter *Interpreter
}

// New creates a new instance of Cint by parsing the provided source string.
// It initializes the lexer, parser, and interpreter for the given source code.
// If parsing errors are encountered, it returns a ParseError containing the errors.
// On success, it returns a pointer to the initialized Cint and a nil error.
func New(source string) (*Cint, error) {
	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return nil, &ParseError{Errors: parser.Errors()}
	}

	interpreter := NewInterpreter(program)

	return &Cint{
		interpreter: interpreter,
	}, nil
}

// Run executes the interpreter associated with the Cint instance.
// It returns an error if the interpreter fails to run.
func (c *Cint) Run() error {
	return c.interpreter.Run()
}

// EnableSingleStep enables single-step execution mode in the underlying interpreter.
// When single-step mode is enabled, the interpreter executes one instruction at a time,
// allowing for fine-grained debugging and inspection of program state after each step.
func (c *Cint) EnableSingleStep() {
	c.interpreter.EnableSingleStep()
}

// DisableSingleStep disables the single-step execution mode in the underlying interpreter.
// When single-step mode is disabled, the interpreter will execute code without pausing
// after each instruction.
func (c *Cint) DisableSingleStep() {
	c.interpreter.DisableSingleStep()
}

// Step executes the next instruction in the interpreter and returns the result as a StepResult.
// It delegates the stepping operation to the underlying interpreter.
func (c *Cint) Step() *StepResult {
	return c.interpreter.Step()
}

// Reset resets the internal interpreter state of the Cint instance.
// This method delegates the reset operation to the underlying interpreter.
func (c *Cint) Reset() {
	c.interpreter.Reset()
}

// ParseError represents an error that occurred during parsing,
// containing a slice of error messages describing the issues encountered.
type ParseError struct {
	Errors []string
}

// Error implements the error interface for ParseError by returning a formatted string
// that lists all parse errors contained in the Errors slice, each on a new line.
func (e *ParseError) Error() string {
	msg := "Parse errors:\n"
	for _, err := range e.Errors {
		msg += "\t" + err + "\n"
	}
	return msg
}
