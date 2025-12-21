package cint

import (
	"fmt"
	"time"
)

// Value represents a runtime value
type Value struct {
	Type  string
	Int   int64
	Float float64
	Str   string
	Ptr   interface{}
}

// Environment stores variables and their values
type Environment struct {
	store map[string]*Value
	outer *Environment
}

// NewEnvironment creates a new environment
func NewEnvironment() *Environment {
	s := make(map[string]*Value)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment creates a new enclosed environment
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a value from the environment
func (e *Environment) Get(name string) (*Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		val, ok = e.outer.Get(name)
	}
	return val, ok
}

// Set sets a value in the environment
func (e *Environment) Set(name string, val *Value) *Value {
	e.store[name] = val
	return val
}

// StepResult represents the result of a single step
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

// Interpreter executes the AST
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

// NewInterpreter creates a new interpreter
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

// EnableSingleStep enables single-stepping mode
func (i *Interpreter) EnableSingleStep() {
	i.stepMode = true
}

// DisableSingleStep disables single-stepping mode
func (i *Interpreter) DisableSingleStep() {
	i.stepMode = false
}

// Run executes the entire program
func (i *Interpreter) Run() error {
	// Execute main function if it exists
	if mainFn, ok := i.functions["main"]; ok {
		i.currentEnv = NewEnclosedEnvironment(i.globals)
		_, err := i.evalFunctionBody(mainFn.Body, i.currentEnv)
		return err
	}
	return fmt.Errorf("no main function found")
}

// Step executes one statement and returns the result
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

// Reset resets the interpreter state for single-stepping
func (i *Interpreter) Reset() {
	i.stepIndex = 0
	i.stepStack = []Statement{}
	i.currentEnv = NewEnvironment()
	i.shouldReturn = false
	i.returnValue = nil
	i.shouldBreak = false
	i.shouldContinue = false
}

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

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Built-in functions
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
}

// processEscapeSequences processes C escape sequences in a string
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
