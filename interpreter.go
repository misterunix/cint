package cint

import (
	"fmt"
	"math"
	"time"
)


// Value represents a dynamically-typed value used by the interpreter.
// It can hold an integer, float, string, or a generic pointer, along with its type as a string.
type Value struct {
	Type  string
	Int   int64
	Float float64
	Str   string
	Ptr   interface{}
}


// Environment represents a variable scope with its own symbol table (store) and an optional
// reference to an outer (enclosing) environment. This structure enables lexical scoping
// and supports nested environments, such as those created by function calls or blocks.
type Environment struct {
	store map[string]*Value
	outer *Environment
}

// NewEnvironment creates and returns a new Environment instance with an empty store.
// The returned Environment has no outer (parent) environment set.
func NewEnvironment() *Environment {
	s := make(map[string]*Value)
	return &Environment{store: s, outer: nil}
}


// NewEnclosedEnvironment creates a new Environment that is enclosed within an outer Environment.
// This allows for nested scopes, where variable lookups will fall back to the outer environment
// if they are not found in the current one.
//
// Parameters:
//   - outer: The enclosing (outer) Environment.
//
// Returns:
//   - A pointer to the newly created enclosed Environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves the Value associated with the given name from the current environment.
// If the name is not found in the current environment, it recursively searches in the outer environments.
// Returns the Value and a boolean indicating whether the name was found.
func (e *Environment) Get(name string) (*Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		val, ok = e.outer.Get(name)
	}
	return val, ok
}


// Set assigns the given Value to the specified name in the Environment's store.
// If the name already exists, its value is overwritten. Returns the assigned Value.
func (e *Environment) Set(name string, val *Value) *Value {
	e.store[name] = val
	return val
}

// StepResult represents the outcome of executing a single step in the interpreter.
// It contains information about the executed statement, the line number, and various
// control flow flags such as Done, Returned, Break, and Continue. It also holds
// the return value (if any) and any error encountered during execution.
type StepResult struct {
	Statement Statement
	Line      int
	Done      bool
	Returned  bool
	ReturnVal *Value
	Break     bool
	Continue  bool
	Error     error
}


// Interpreter represents the core execution context for the C interpreter.
// It maintains the current program, global environment, function declarations,
// built-in functions, and manages execution state such as stepping, control flow,
// and the current environment for statement execution.
type Interpreter struct {
	program    *Program
	globals    *Environment
	functions  map[string]*FunctionDecl
	builtins   map[string]func([]Expression, *Environment) (*Value, error)
	stepMode   bool
	stepIndex  int
	stepStack  []Statement // Stack of statements to execute
	currentEnv *Environment

	// Control flow
	shouldReturn   bool
	returnValue    *Value
	shouldBreak    bool
	shouldContinue bool
}


// NewInterpreter creates and initializes a new Interpreter instance for the given Program.
// It sets up the global environment, registers built-in functions, and collects all function
// declarations from the program for later use.
//
// Parameters:
//   - program: The Program to be interpreted.
//
// Returns:
//   - A pointer to the initialized Interpreter.
func NewInterpreter(program *Program) *Interpreter {
	interp := &Interpreter{
		program:   program,
		globals:   NewEnvironment(),
		functions: make(map[string]*FunctionDecl),
		builtins:  make(map[string]func([]Expression, *Environment) (*Value, error)),
		stepStack: []Statement{},
	}

	// Register built-in functions
	interp.registerBuiltins()

	// Extract function declarations
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*FunctionDecl); ok {
			interp.functions[fn.Name] = fn
		}
	}

	return interp
}


// EnableSingleStep enables single-step execution mode in the interpreter.
// When single-step mode is active, the interpreter executes one instruction at a time,
// allowing for step-by-step debugging or inspection.
func (i *Interpreter) EnableSingleStep() {
	i.stepMode = true
}


// DisableSingleStep disables the single-step execution mode in the interpreter.
// When called, the interpreter will no longer pause after executing each instruction.
func (i *Interpreter) DisableSingleStep() {
	i.stepMode = false
}


// Run executes the "main" function of the interpreter if it exists.
// It sets up a new environment enclosed within the global environment,
// evaluates the body of the main function, and returns any error encountered.
// If no "main" function is found, it returns an error indicating this.
func (i *Interpreter) Run() error {
	// Execute main function if it exists
	if mainFn, ok := i.functions["main"]; ok {
		i.currentEnv = NewEnclosedEnvironment(i.globals)
		_, err := i.evalFunctionBody(mainFn.Body, i.currentEnv)
		return err
	}
	return fmt.Errorf("no main function found")
}


// Step executes the next statement in single-step mode for the interpreter.
// It initializes the step stack with the main function's statements on the first call.
// Returns a StepResult containing the executed statement, any error encountered,
// and flags indicating if execution is done, if a return, break, or continue was triggered,
// and the return value if applicable. If single-step mode is not enabled or no main function
// is found, returns an appropriate error in the StepResult.
func (i *Interpreter) Step() *StepResult {
	if !i.stepMode {
		return &StepResult{Error: fmt.Errorf("single-step mode not enabled")}
	}

	// Initialize on first step
	if i.stepIndex == 0 && len(i.stepStack) == 0 {
		if mainFn, ok := i.functions["main"]; ok {
			if mainFn.Body != nil {
				i.currentEnv = NewEnclosedEnvironment(i.globals)
				i.stepStack = append(i.stepStack, mainFn.Body.Statements...)
			}
		} else {
			return &StepResult{Error: fmt.Errorf("no main function found"), Done: true}
		}
	}

	// Check if we're done
	if i.stepIndex >= len(i.stepStack) {
		return &StepResult{Done: true}
	}

	stmt := i.stepStack[i.stepIndex]
	i.stepIndex++

	// Evaluate the statement
	err := i.evalStatement(stmt, i.currentEnv)

	result := &StepResult{
		Statement: stmt,
		Done:      i.stepIndex >= len(i.stepStack) || i.shouldReturn,
		Returned:  i.shouldReturn,
		ReturnVal: i.returnValue,
		Break:     i.shouldBreak,
		Continue:  i.shouldContinue,
		Error:     err,
	}

	return result
}


// Reset reinitializes the Interpreter to its default state, clearing the step stack,
// resetting the step index, creating a new environment, and clearing any control flow
// or return flags. This prepares the Interpreter for a fresh execution.
func (i *Interpreter) Reset() {
	i.stepIndex = 0
	i.stepStack = []Statement{}
	i.currentEnv = NewEnvironment()
	i.shouldReturn = false
	i.returnValue = nil
	i.shouldBreak = false
	i.shouldContinue = false
}

// evalStatement evaluates a given Statement node within the provided Environment.
// It dispatches the evaluation based on the concrete type of the Statement, handling
// variable declarations, expressions, control flow statements (if, while, for, block),
// and flow control (return, break, continue). The method updates interpreter state
// flags (shouldReturn, shouldBreak, shouldContinue) as needed and returns any error
// encountered during evaluation.
func (i *Interpreter) evalStatement(stmt Statement, env *Environment) error {
	switch node := stmt.(type) {
	case *VarDecl:
		return i.evalVarDecl(node, env)
	case *ExpressionStatement:
		_, err := i.evalExpression(node.Expression, env)
		return err
	case *ReturnStatement:
		if node.ReturnValue != nil {
			val, err := i.evalExpression(node.ReturnValue, env)
			if err != nil {
				return err
			}
			i.returnValue = val
		}
		i.shouldReturn = true
		return nil
	case *IfStatement:
		return i.evalIfStatement(node, env)
	case *WhileStatement:
		return i.evalWhileStatement(node, env)
	case *ForStatement:
		return i.evalForStatement(node, env)
	case *BlockStatement:
		return i.evalBlockStatement(node, env)
	case *BreakStatement:
		i.shouldBreak = true
		return nil
	case *ContinueStatement:
		i.shouldContinue = true
		return nil
	}
	return nil
}

// evalVarDecl evaluates a variable declaration node within the given environment.
// If the variable declaration includes an initial value, it evaluates the expression
// and assigns the result to the variable. Otherwise, it initializes the variable with
// a default value (zero for integers). The variable is then set in the environment.
// Returns an error if evaluation of the initial value fails.
func (i *Interpreter) evalVarDecl(node *VarDecl, env *Environment) error {
	var val *Value

	if node.Value != nil {
		var err error
		val, err = i.evalExpression(node.Value, env)
		if err != nil {
			return err
		}
	} else {
		// Default initialization
		val = &Value{Type: node.Type, Int: 0}
	}

	env.Set(node.Name, val)
	return nil
}

// evalBlockStatement evaluates each statement within the provided BlockStatement
// in the given Environment. It processes statements sequentially, stopping early
// if a return, break, or continue condition is triggered. Returns an error if
// any statement evaluation fails, otherwise returns nil.
func (i *Interpreter) evalBlockStatement(block *BlockStatement, env *Environment) error {
	for _, stmt := range block.Statements {
		if err := i.evalStatement(stmt, env); err != nil {
			return err
		}
		if i.shouldReturn || i.shouldBreak || i.shouldContinue {
			break
		}
	}
	return nil
}

// evalIfStatement evaluates an IfStatement node within the given environment.
// It first evaluates the condition expression. If the condition is truthy,
// it evaluates and returns the result of the consequence block. If the condition
// is falsy and an alternative block exists, it evaluates and returns the result
// of the alternative block. If neither block is executed, it returns nil.
// Returns an error if evaluating the condition or any block fails.
func (i *Interpreter) evalIfStatement(node *IfStatement, env *Environment) error {
	condition, err := i.evalExpression(node.Condition, env)
	if err != nil {
		return err
	}

	if i.isTruthy(condition) {
		return i.evalBlockStatement(node.Consequence, env)
	} else if node.Alternative != nil {
		return i.evalBlockStatement(node.Alternative, env)
	}

	return nil
}

// evalWhileStatement evaluates a WhileStatement node within the given environment.
// It repeatedly evaluates the loop's condition and executes the loop body as long as the condition is truthy.
// The method handles control flow statements such as return, break, and continue by checking the interpreter's flags:
// - If shouldReturn is set, the loop breaks to allow return propagation.
// - If shouldBreak is set, the flag is reset and the loop breaks.
// - If shouldContinue is set, the flag is reset and the loop continues to the next iteration.
// Returns an error if evaluating the condition or body fails.
func (i *Interpreter) evalWhileStatement(node *WhileStatement, env *Environment) error {
	for {
		condition, err := i.evalExpression(node.Condition, env)
		if err != nil {
			return err
		}

		if !i.isTruthy(condition) {
			break
		}

		if err := i.evalBlockStatement(node.Body, env); err != nil {
			return err
		}

		if i.shouldReturn {
			break
		}
		if i.shouldBreak {
			i.shouldBreak = false
			break
		}
		if i.shouldContinue {
			i.shouldContinue = false
			continue
		}
	}
	return nil
}

// evalForStatement evaluates a ForStatement node within the interpreter.
// It creates a new scope for the loop, initializes any loop variables,
// checks the loop condition, executes the loop body, and handles post-iteration
// expressions. The method also manages control flow statements such as break,
// continue, and return within the loop. It returns an error if any statement
// or expression evaluation fails.
func (i *Interpreter) evalForStatement(node *ForStatement, env *Environment) error {
	// Create new scope for the loop
	loopEnv := NewEnclosedEnvironment(env)

	// Initialize
	if node.Init != nil {
		if err := i.evalStatement(node.Init, loopEnv); err != nil {
			return err
		}
	}

	for {
		// Check condition
		if node.Condition != nil {
			condition, err := i.evalExpression(node.Condition, loopEnv)
			if err != nil {
				return err
			}
			if !i.isTruthy(condition) {
				break
			}
		}

		// Execute body
		if err := i.evalBlockStatement(node.Body, loopEnv); err != nil {
			return err
		}

		if i.shouldReturn {
			break
		}
		if i.shouldBreak {
			i.shouldBreak = false
			break
		}
		if i.shouldContinue {
			i.shouldContinue = false
		}

		// Execute post
		if node.Post != nil {
			if _, err := i.evalExpression(node.Post, loopEnv); err != nil {
				return err
			}
		}
	}

	return nil
}

// evalFunctionBody evaluates the body of a function represented by the given BlockStatement
// within the provided Environment. It resets the interpreter's return state before execution.
// If a return statement is encountered during evaluation, the corresponding value is returned.
// If no explicit return is found, a default integer value of 0 is returned.
// Returns an error if evaluation of the block fails.
func (i *Interpreter) evalFunctionBody(block *BlockStatement, env *Environment) (*Value, error) {
	i.shouldReturn = false
	i.returnValue = nil

	if err := i.evalBlockStatement(block, env); err != nil {
		return nil, err
	}

	if i.returnValue != nil {
		return i.returnValue, nil
	}

	return &Value{Type: "int", Int: 0}, nil
}

// evalExpression evaluates the given Expression node within the provided Environment.
// It dispatches evaluation based on the concrete type of the Expression, handling literals,
// identifiers, prefix/postfix/infix expressions, assignments, function calls, and conditionals.
// Returns the resulting Value and an error if evaluation fails (e.g., undefined variable or unknown expression type).
func (i *Interpreter) evalExpression(expr Expression, env *Environment) (*Value, error) {
	switch node := expr.(type) {
	case *IntegerLiteral:
		return &Value{Type: "int", Int: node.Value}, nil
	case *FloatLiteral:
		return &Value{Type: "float", Float: node.Value}, nil
	case *StringLiteral:
		return &Value{Type: "string", Str: node.Value}, nil
	case *CharLiteral:
		return &Value{Type: "char", Int: int64(node.Value)}, nil
	case *Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", node.Value)
		}
		return val, nil
	case *PrefixExpression:
		return i.evalPrefixExpression(node, env)
	case *PostfixExpression:
		return i.evalPostfixExpression(node, env)
	case *InfixExpression:
		return i.evalInfixExpression(node, env)
	case *AssignmentExpression:
		return i.evalAssignmentExpression(node, env)
	case *CallExpression:
		return i.evalCallExpression(node, env)
	case *ConditionalExpression:
		return i.evalConditionalExpression(node, env)
	}
	return nil, fmt.Errorf("unknown expression type")
}

// evalPrefixExpression evaluates a prefix expression node within the given environment.
// It supports the following prefix operators:
//   - "-"  : Negates the value (supports both int and float types).
//   - "!"  : Logical NOT, returns 1 if the value is falsy, 0 otherwise.
//   - "~"  : Bitwise NOT, applies only to int values.
//   - "++" : Pre-increment, increments the value of an identifier before returning it.
//   - "--" : Pre-decrement, decrements the value of an identifier before returning it.
// Returns the evaluated Value or an error if the operator is unknown or evaluation fails.
func (i *Interpreter) evalPrefixExpression(node *PrefixExpression, env *Environment) (*Value, error) {
	right, err := i.evalExpression(node.Right, env)
	if err != nil {
		return nil, err
	}

	switch node.Operator {
	case "-":
		if right.Type == "float" {
			return &Value{Type: "float", Float: -right.Float}, nil
		}
		return &Value{Type: "int", Int: -right.Int}, nil
	case "!":
		return &Value{Type: "int", Int: boolToInt(!i.isTruthy(right))}, nil
	case "~":
		return &Value{Type: "int", Int: ^right.Int}, nil
	case "++":
		// Pre-increment
		if ident, ok := node.Right.(*Identifier); ok {
			val, _ := env.Get(ident.Value)
			val.Int++
			return val, nil
		}
	case "--":
		// Pre-decrement
		if ident, ok := node.Right.(*Identifier); ok {
			val, _ := env.Get(ident.Value)
			val.Int--
			return val, nil
		}
	}

	return nil, fmt.Errorf("unknown prefix operator: %s", node.Operator)
}

// evalPostfixExpression evaluates a postfix expression (such as increment '++' or decrement '--')
// for the given AST node and environment. It returns the value of the expression before the postfix
// operation is applied, as per C-like semantics. If the operator is not recognized, an error is returned.
// Only identifiers can be incremented or decremented; otherwise, the function returns an error.
func (i *Interpreter) evalPostfixExpression(node *PostfixExpression, env *Environment) (*Value, error) {
	left, err := i.evalExpression(node.Left, env)
	if err != nil {
		return nil, err
	}

	oldValue := &Value{Type: left.Type, Int: left.Int, Float: left.Float}

	switch node.Operator {
	case "++":
		if ident, ok := node.Left.(*Identifier); ok {
			val, _ := env.Get(ident.Value)
			val.Int++
		}
		return oldValue, nil
	case "--":
		if ident, ok := node.Left.(*Identifier); ok {
			val, _ := env.Get(ident.Value)
			val.Int--
		}
		return oldValue, nil
	}

	return nil, fmt.Errorf("unknown postfix operator: %s", node.Operator)
}

// evalInfixExpression evaluates an infix expression node within the given environment.
// It supports operations on both integer and float types, including arithmetic, comparison,
// logical, and bitwise operators. The function first evaluates the left and right operands,
// then determines the operation to perform based on the operator and operand types.
// Returns the resulting Value and an error if any occurs (e.g., division by zero or unknown operator).
//
// Supported operators:
//   - Arithmetic: +, -, *, /, %
//   - Comparison: <, >, <=, >=, ==, !=
//   - Logical: &&, ||
//   - Bitwise: &, |, ^, <<, >>
// For float operands, only arithmetic and comparison operators are supported.
func (i *Interpreter) evalInfixExpression(node *InfixExpression, env *Environment) (*Value, error) {
	left, err := i.evalExpression(node.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := i.evalExpression(node.Right, env)
	if err != nil {
		return nil, err
	}

	// Handle float operations
	if left.Type == "float" || right.Type == "float" {
		leftF := left.Float
		if left.Type != "float" {
			leftF = float64(left.Int)
		}
		rightF := right.Float
		if right.Type != "float" {
			rightF = float64(right.Int)
		}

		switch node.Operator {
		case "+":
			return &Value{Type: "float", Float: leftF + rightF}, nil
		case "-":
			return &Value{Type: "float", Float: leftF - rightF}, nil
		case "*":
			return &Value{Type: "float", Float: leftF * rightF}, nil
		case "/":
			return &Value{Type: "float", Float: leftF / rightF}, nil
		case "<":
			return &Value{Type: "int", Int: boolToInt(leftF < rightF)}, nil
		case ">":
			return &Value{Type: "int", Int: boolToInt(leftF > rightF)}, nil
		case "<=":
			return &Value{Type: "int", Int: boolToInt(leftF <= rightF)}, nil
		case ">=":
			return &Value{Type: "int", Int: boolToInt(leftF >= rightF)}, nil
		case "==":
			return &Value{Type: "int", Int: boolToInt(leftF == rightF)}, nil
		case "!=":
			return &Value{Type: "int", Int: boolToInt(leftF != rightF)}, nil
		}
	}

	// Integer operations
	switch node.Operator {
	case "+":
		return &Value{Type: "int", Int: left.Int + right.Int}, nil
	case "-":
		return &Value{Type: "int", Int: left.Int - right.Int}, nil
	case "*":
		return &Value{Type: "int", Int: left.Int * right.Int}, nil
	case "/":
		if right.Int == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &Value{Type: "int", Int: left.Int / right.Int}, nil
	case "%":
		return &Value{Type: "int", Int: left.Int % right.Int}, nil
	case "<":
		return &Value{Type: "int", Int: boolToInt(left.Int < right.Int)}, nil
	case ">":
		return &Value{Type: "int", Int: boolToInt(left.Int > right.Int)}, nil
	case "<=":
		return &Value{Type: "int", Int: boolToInt(left.Int <= right.Int)}, nil
	case ">=":
		return &Value{Type: "int", Int: boolToInt(left.Int >= right.Int)}, nil
	case "==":
		return &Value{Type: "int", Int: boolToInt(left.Int == right.Int)}, nil
	case "!=":
		return &Value{Type: "int", Int: boolToInt(left.Int != right.Int)}, nil
	case "&&":
		return &Value{Type: "int", Int: boolToInt(i.isTruthy(left) && i.isTruthy(right))}, nil
	case "||":
		return &Value{Type: "int", Int: boolToInt(i.isTruthy(left) || i.isTruthy(right))}, nil
	case "&":
		return &Value{Type: "int", Int: left.Int & right.Int}, nil
	case "|":
		return &Value{Type: "int", Int: left.Int | right.Int}, nil
	case "^":
		return &Value{Type: "int", Int: left.Int ^ right.Int}, nil
	case "<<":
		return &Value{Type: "int", Int: left.Int << uint(right.Int)}, nil
	case ">>":
		return &Value{Type: "int", Int: left.Int >> uint(right.Int)}, nil
	}

	return nil, fmt.Errorf("unknown infix operator: %s", node.Operator)
}

// evalAssignmentExpression evaluates an assignment expression node within the given environment.
// It supports both simple assignments (e.g., x = 5) and compound assignments (e.g., x += 2).
// The function first evaluates the right-hand side expression. If the left-hand side is an identifier,
// it performs the assignment or compound operation, updating the environment accordingly.
// Returns the resulting value of the assignment or an error if the operation is invalid or the variable is undefined.
func (i *Interpreter) evalAssignmentExpression(node *AssignmentExpression, env *Environment) (*Value, error) {
	right, err := i.evalExpression(node.Right, env)
	if err != nil {
		return nil, err
	}

	if ident, ok := node.Left.(*Identifier); ok {
		if node.Operator == "=" {
			env.Set(ident.Value, right)
			return right, nil
		}

		// Compound assignment
		left, ok := env.Get(ident.Value)
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", ident.Value)
		}

		var result *Value
		switch node.Operator {
		case "+=":
			result = &Value{Type: left.Type, Int: left.Int + right.Int}
		case "-=":
			result = &Value{Type: left.Type, Int: left.Int - right.Int}
		case "*=":
			result = &Value{Type: left.Type, Int: left.Int * right.Int}
		case "/=":
			result = &Value{Type: left.Type, Int: left.Int / right.Int}
		case "%=":
			result = &Value{Type: left.Type, Int: left.Int % right.Int}
		case "&=":
			result = &Value{Type: left.Type, Int: left.Int & right.Int}
		case "|=":
			result = &Value{Type: left.Type, Int: left.Int | right.Int}
		case "^=":
			result = &Value{Type: left.Type, Int: left.Int ^ right.Int}
		case "<<=":
			result = &Value{Type: left.Type, Int: left.Int << uint(right.Int)}
		case ">>=":
			result = &Value{Type: left.Type, Int: left.Int >> uint(right.Int)}
		default:
			return nil, fmt.Errorf("unknown assignment operator: %s", node.Operator)
		}

		env.Set(ident.Value, result)
		return result, nil
	}

	return nil, fmt.Errorf("invalid assignment target")
}

// evalCallExpression evaluates a function call expression within the interpreter.
// It first determines the function name from the provided node. If the function is a built-in,
// it invokes the corresponding built-in implementation. If the function is user-defined, it creates
// a new environment, binds the arguments to the function parameters, and executes the function body.
// The function also preserves and restores the interpreter's return state to handle nested returns correctly.
// Returns the result of the function call or an error if the function is undefined or evaluation fails.
func (i *Interpreter) evalCallExpression(node *CallExpression, env *Environment) (*Value, error) {
	funcName := ""
	if ident, ok := node.Function.(*Identifier); ok {
		funcName = ident.Value
	} else {
		return nil, fmt.Errorf("invalid function call")
	}

	// Check for built-in functions
	if builtin, ok := i.builtins[funcName]; ok {
		return builtin(node.Arguments, env)
	}

	// Check for user-defined functions
	if fn, ok := i.functions[funcName]; ok {
		// Save current return state
		savedShouldReturn := i.shouldReturn
		savedReturnValue := i.returnValue

		// Create new environment for function
		fnEnv := NewEnclosedEnvironment(i.globals)

		// Bind parameters
		for idx, param := range fn.Parameters {
			if idx < len(node.Arguments) {
				val, err := i.evalExpression(node.Arguments[idx], env)
				if err != nil {
					return nil, err
				}
				fnEnv.Set(param.Name, val)
			}
		}

		// Execute function body
		result, err := i.evalFunctionBody(fn.Body, fnEnv)

		// Restore return state
		i.shouldReturn = savedShouldReturn
		i.returnValue = savedReturnValue

		return result, err
	}

	return nil, fmt.Errorf("undefined function: %s", funcName)
}

// evalConditionalExpression evaluates a conditional (ternary) expression node within the interpreter.
// It first evaluates the condition expression. If the condition is truthy, it evaluates and returns
// the consequence expression; otherwise, it evaluates and returns the alternative expression.
// Returns the resulting Value or an error if evaluation fails.
func (i *Interpreter) evalConditionalExpression(node *ConditionalExpression, env *Environment) (*Value, error) {
	condition, err := i.evalExpression(node.Condition, env)
	if err != nil {
		return nil, err
	}

	if i.isTruthy(condition) {
		return i.evalExpression(node.Consequence, env)
	}
	return i.evalExpression(node.Alternative, env)
}

func (i *Interpreter) isTruthy(val *Value) bool {
	if val.Type == "float" {
		return val.Float != 0.0
	}
	return val.Int != 0
}

// boolToInt converts a boolean value to its integer representation.
// It returns 1 if the input is true, and 0 if the input is false.
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}


// registerBuiltins registers a set of built-in functions into the interpreter's environment.
// These built-ins include:
//   - printf: Prints formatted output to stdout, similar to C's printf.
//   - sleep: Pauses execution for a specified number of milliseconds.
//   - putchar: Prints a single character to stdout.
//   - sqrt: Returns the square root of a number.
//   - pow: Raises a number to the power of another.
//   - sin, cos, tan: Trigonometric functions (sine, cosine, tangent).
//   - abs: Returns the absolute value of a number.
//   - floor: Rounds a number down to the nearest integer.
//   - ceil: Rounds a number up to the nearest integer.
//   - log: Returns the natural logarithm of a number.
//   - log10: Returns the base-10 logarithm of a number.
//   - exp: Returns e raised to the power of a number.
// Each built-in function is added to the interpreter's builtins map and can be invoked from interpreted code.
func (i *Interpreter) registerBuiltins() {
	// printf
	i.builtins["printf"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return &Value{Type: "int", Int: 0}, nil
		}

		formatVal, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}

		format := processEscapeSequences(formatVal.Str)
		argVals := []interface{}{}

		for idx := 1; idx < len(args); idx++ {
			val, err := i.evalExpression(args[idx], env)
			if err != nil {
				return nil, err
			}

			if val.Type == "float" {
				argVals = append(argVals, val.Float)
			} else if val.Type == "string" {
				argVals = append(argVals, val.Str)
			} else {
				argVals = append(argVals, val.Int)
			}
		}

		fmt.Printf(format, argVals...)
		return &Value{Type: "int", Int: 0}, nil
	}

	// sleep - millisecond resolution
	i.builtins["sleep"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return &Value{Type: "int", Int: 0}, nil
		}

		msVal, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}

		time.Sleep(time.Duration(msVal.Int) * time.Millisecond)
		return &Value{Type: "int", Int: 0}, nil
	}

	// putchar
	i.builtins["putchar"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return &Value{Type: "int", Int: 0}, nil
		}

		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%c", byte(val.Int))
		return &Value{Type: "int", Int: val.Int}, nil
	}

	// Floating point math functions

	// sqrt - square root
	i.builtins["sqrt"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("sqrt requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Sqrt(f)}, nil
	}

	// pow - power (x^y)
	i.builtins["pow"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("pow requires 2 arguments")
		}
		base, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		exp, err := i.evalExpression(args[1], env)
		if err != nil {
			return nil, err
		}
		var baseF, expF float64
		if base.Type == "float" {
			baseF = base.Float
		} else {
			baseF = float64(base.Int)
		}
		if exp.Type == "float" {
			expF = exp.Float
		} else {
			expF = float64(exp.Int)
		}
		return &Value{Type: "float", Float: math.Pow(baseF, expF)}, nil
	}

	// sin - sine
	i.builtins["sin"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("sin requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Sin(f)}, nil
	}

	// cos - cosine
	i.builtins["cos"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("cos requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Cos(f)}, nil
	}

	// tan - tangent
	i.builtins["tan"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("tan requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Tan(f)}, nil
	}

	// abs - absolute value
	i.builtins["abs"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("abs requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		if val.Type == "float" {
			return &Value{Type: "float", Float: math.Abs(val.Float)}, nil
		}
		if val.Int < 0 {
			return &Value{Type: "int", Int: -val.Int}, nil
		}
		return &Value{Type: "int", Int: val.Int}, nil
	}

	// floor - round down
	i.builtins["floor"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("floor requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Floor(f)}, nil
	}

	// ceil - round up
	i.builtins["ceil"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("ceil requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Ceil(f)}, nil
	}

	// log - natural logarithm
	i.builtins["log"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("log requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Log(f)}, nil
	}

	// log10 - logarithm base 10
	i.builtins["log10"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("log10 requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Log10(f)}, nil
	}

	// exp - exponential (e^x)
	i.builtins["exp"] = func(args []Expression, env *Environment) (*Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("exp requires 1 argument")
		}
		val, err := i.evalExpression(args[0], env)
		if err != nil {
			return nil, err
		}
		var f float64
		if val.Type == "float" {
			f = val.Float
		} else {
			f = float64(val.Int)
		}
		return &Value{Type: "float", Float: math.Exp(f)}, nil
	}
}


// processEscapeSequences takes a string containing C-style escape sequences
// (such as \n, \t, \r, \\, \", and \0) and returns a new string with those
// sequences replaced by their corresponding characters. Unrecognized escape
// sequences are left as-is (the backslash and following character are both
// included in the result).
func processEscapeSequences(s string) string {
	result := ""
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result += "\n"
			case 't':
				result += "\t"
			case 'r':
				result += "\r"
			case '\\':
				result += "\\"
			case '"':
				result += "\""
			case '0':
				result += "\x00"
			default:
				result += string(s[i])
				result += string(s[i+1])
			}
			i += 2
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
