package cint

// Cint provides the main interface to the C interpreter
type Cint struct {
	interpreter *Interpreter
}

// New creates a new C interpreter instance
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

// Run executes the C program
func (c *Cint) Run() error {
	return c.interpreter.Run()
}

// EnableSingleStep enables single-stepping mode
func (c *Cint) EnableSingleStep() {
	c.interpreter.EnableSingleStep()
}

// DisableSingleStep disables single-stepping mode
func (c *Cint) DisableSingleStep() {
	c.interpreter.DisableSingleStep()
}

// Step executes one statement in single-step mode
func (c *Cint) Step() *StepResult {
	return c.interpreter.Step()
}

// Reset resets the interpreter state
func (c *Cint) Reset() {
	c.interpreter.Reset()
}

// ParseError represents parsing errors
type ParseError struct {
	Errors []string
}

func (e *ParseError) Error() string {
	msg := "Parse errors:\n"
	for _, err := range e.Errors {
		msg += "\t" + err + "\n"
	}
	return msg
}
