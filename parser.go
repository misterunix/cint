package cint

import (
	"fmt"
	"strconv"
)


// Parser represents a recursive descent parser for the C language.
// It maintains the current and next tokens, a reference to the lexer,
// and a list of parsing errors encountered during processing.
type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}


// NewParser creates and returns a new Parser instance using the provided Lexer.
// It initializes the parser by advancing the lexer twice to set up the current and peek tokens.
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances the parser to the next token by updating the current and peek tokens.
// It sets curToken to the current peekToken, and then fetches the next token from the lexer
// to update peekToken. This method is typically used to iterate through the input tokens
// during parsing.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token's type matches the provided TokenType t.
// It returns true if the types are equal, otherwise false.
func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the next token (peekToken) is of the specified TokenType t.
// It returns true if the type matches, otherwise false.
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the expected type 't'.
// If it is, the parser advances to the next token and returns true.
// Otherwise, it records a peek error and returns false.
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError records an error message when the next token does not match the expected TokenType.
// It appends a formatted error message to the parser's error list, including details about the
// expected and actual token types, their string representations, the literal value, and the line number.
func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %v (type %d), got %v (type %d) '%s' instead at line %d",
		t, t, p.peekToken.Type, p.peekToken.Type, p.peekToken.Literal, p.peekToken.Line)
	p.errors = append(p.errors, msg)
}

// Errors returns a slice of error messages encountered during parsing.
func (p *Parser) Errors() []string {
	return p.errors
}


// ParseProgram parses the entire input and constructs a Program AST node.
// It iterates through all tokens until EOF, parsing each statement and
// appending it to the Program's Statements slice. Returns the fully
// constructed Program node.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for !p.curTokenIs(EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement parses the current token and returns the corresponding Statement node.
// It determines the type of statement based on the current token, handling declarations,
// control flow statements (return, if, while, for, break, continue), block statements,
// and expression statements. The method delegates parsing to specialized functions
// depending on the token type.
func (p *Parser) parseStatement() Statement {
	// Check for type keywords (variable or function declaration)
	if p.isTypeKeyword(p.curToken.Type) {
		return p.parseDeclaration()
	}

	switch p.curToken.Type {
	case RETURN:
		return p.parseReturnStatement()
	case IF:
		return p.parseIfStatement()
	case WHILE:
		return p.parseWhileStatement()
	case FOR:
		return p.parseForStatement()
	case BREAK:
		return p.parseBreakStatement()
	case CONTINUE:
		return p.parseContinueStatement()
	case LBRACE:
		return p.parseBlockStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// isTypeKeyword checks if the given TokenType represents a C type keyword,
// such as int, char, float, double, void, long, short, unsigned, or signed.
func (p *Parser) isTypeKeyword(t TokenType) bool {
	return t == INT_KW || t == CHAR_KW || t == FLOAT_KW || t == DOUBLE ||
		t == VOID || t == LONG || t == SHORT || t == UNSIGNED || t == SIGNED
}

// parseType parses and returns a type name from the current token stream.
// It starts with the current token's literal value and appends a '*' for each
// subsequent pointer indicator (STAR token) encountered, advancing the token
// stream as it goes. The resulting string represents the parsed type,
// including any pointer indirection.
func (p *Parser) parseType() string {
	typ := p.curToken.Literal
	p.nextToken()

	// Handle pointer types
	for p.curTokenIs(STAR) {
		typ += "*"
		p.nextToken()
	}

	return typ
}

// parseDeclaration parses a declaration statement in the source code.
// It first attempts to parse a type, then checks if the next token is an identifier.
// If the identifier is followed by a left parenthesis, it is treated as a function declaration
// and delegated to parseFunctionDecl. Otherwise, it is treated as a variable declaration
// and delegated to parseVarDecl. Returns a Statement representing the parsed declaration,
// or nil if the declaration is invalid.
func (p *Parser) parseDeclaration() Statement {
	typ := p.parseType()

	if !p.curTokenIs(IDENT) {
		return nil
	}

	name := p.curToken.Literal
	nameToken := p.curToken

	// Check if it's a function declaration (peek ahead)
	if p.peekTokenIs(LPAREN) {
		// Don't consume identifier yet, let parseFunctionDecl handle it
		return p.parseFunctionDecl(typ, name, nameToken)
	}

	// It's a variable declaration
	p.nextToken() // consume the identifier
	return p.parseVarDecl(typ, name, nameToken)
}

// parseFunctionDecl parses a function declaration starting from the given return type, function name, and token.
// It expects the next token to be a left parenthesis '(', followed by zero or more parameter declarations,
// and a closing right parenthesis ')'. Each parameter consists of a type and an optional name.
// After the parameter list, it checks for either a function body (enclosed in braces) or a semicolon
// indicating a function prototype. Returns a pointer to the constructed FunctionDecl, or nil if parsing fails.
func (p *Parser) parseFunctionDecl(returnType, name string, token Token) *FunctionDecl {
	fn := &FunctionDecl{
		Token:      token,
		ReturnType: returnType,
		Name:       name,
		Parameters: []*Parameter{},
	}

	// Expect ( and move to it
	if !p.expectPeek(LPAREN) {
		// Debug
		// fmt.Printf("parseFunctionDecl: expected LPAREN, cur=%v peek=%v\n", p.curToken, p.peekToken)
		return nil
	}

	// Now at (, move to next token
	p.nextToken()

	// Parse parameters
	if !p.curTokenIs(RPAREN) {
		for {
			if !p.isTypeKeyword(p.curToken.Type) {
				break
			}

			paramType := p.parseType()
			paramName := ""
			if p.curTokenIs(IDENT) {
				paramName = p.curToken.Literal
				p.nextToken()
			}

			fn.Parameters = append(fn.Parameters, &Parameter{
				Type: paramType,
				Name: paramName,
			})

			if !p.curTokenIs(COMMA) {
				break
			}
			p.nextToken()
		}
	}

	// Should be at )
	if !p.curTokenIs(RPAREN) {
		p.peekError(RPAREN)
		return nil
	}

	// Move past )
	p.nextToken()

	// Check for function body or just declaration
	if p.curTokenIs(LBRACE) {
		fn.Body = p.parseBlockStatement()
	} else if p.curTokenIs(SEMICOLON) {
		// Function declaration only
	}

	return fn
}

// parseVarDecl parses a variable declaration statement, including optional array
// notation and initialization. It constructs and returns a VarDecl node with the
// provided type, name, and token. If the declaration includes array brackets,
// the type is modified to indicate an array. If an assignment is present, the
// initialization expression is parsed and attached. The function also ensures
// the statement is properly terminated with a semicolon.
//
// Parameters:
//   typ   - the type of the variable being declared
//   name  - the name of the variable
//   token - the token representing the start of the declaration
//
// Returns:
//   *VarDecl - the parsed variable declaration node
func (p *Parser) parseVarDecl(typ, name string, token Token) *VarDecl {
	vd := &VarDecl{
		Token: token,
		Type:  typ,
		Name:  name,
	}

	// Check for array declaration
	if p.curTokenIs(LBRACKET) {
		p.nextToken()
		if !p.curTokenIs(RBRACKET) {
			p.nextToken() // skip size for now
		}
		p.nextToken() // consume ]
		vd.Type += "[]"
	}

	// Check for initialization
	if p.curTokenIs(ASSIGN) {
		p.nextToken()
		vd.Value = p.parseExpression(LOWEST)
	}

	if p.curTokenIs(SEMICOLON) {
		// Already at semicolon, good
	} else if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return vd
}

// parseBlockStatement parses a block statement, which is a series of statements enclosed by braces.
// It advances the parser to the next token, collects all statements until it encounters a closing brace (RBRACE)
// or the end of file (EOF), and returns a BlockStatement containing the parsed statements.
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	p.nextToken()

	for !p.curTokenIs(RBRACE) && !p.curTokenIs(EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseReturnStatement parses a return statement in the source code.
// It creates a new ReturnStatement node, advances the token stream,
// and parses the return value expression if present. It also handles
// optional semicolons after the return value for statement termination.
// Returns the constructed *ReturnStatement AST node.
func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}
	p.nextToken()

	if !p.curTokenIs(SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.curTokenIs(SEMICOLON) {
		// good
	} else if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseIfStatement parses an 'if' statement from the current token stream.
// It expects the following syntax: 'if (condition) { consequence } [else { alternative }]'
// The method advances the parser through the tokens, building an IfStatement AST node
// with the parsed condition, consequence block, and optional alternative block.
// Returns a pointer to the constructed IfStatement, or nil if parsing fails at any step.
func (p *Parser) parseIfStatement() *IfStatement {
	stmt := &IfStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(ELSE) {
		p.nextToken()

		if p.peekTokenIs(LBRACE) {
			p.nextToken()
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseWhileStatement parses a 'while' statement from the current token stream.
// It expects the following syntax: 'while (condition) { body }'.
// The method advances the parser through the tokens, parses the condition expression
// inside parentheses, and then parses the block statement as the loop body.
// Returns a pointer to a WhileStatement node representing the parsed 'while' statement,
// or nil if the expected tokens are not found in the correct order.
func (p *Parser) parseWhileStatement() *WhileStatement {
	stmt := &WhileStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement parses a 'for' statement from the current token stream.
// It expects the following syntax: for (init; condition; post) { body }.
// The function parses the initialization statement, condition expression, and post expression,
// as well as the block statement that forms the body of the loop.
// Returns a pointer to a ForStatement AST node, or nil if parsing fails at any stage.
func (p *Parser) parseForStatement() *ForStatement {
	stmt := &ForStatement{Token: p.curToken}

	if !p.expectPeek(LPAREN) {
		return nil
	}

	p.nextToken()

	// Parse initialization
	if !p.curTokenIs(SEMICOLON) {
		stmt.Init = p.parseStatement()
	}

	if p.curTokenIs(SEMICOLON) {
		p.nextToken()
	}

	// Parse condition
	if !p.curTokenIs(SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(SEMICOLON) {
		return nil
	}

	p.nextToken()

	// Parse post expression
	if !p.curTokenIs(RPAREN) {
		stmt.Post = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(RPAREN) {
		return nil
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseBreakStatement parses a 'break' statement from the current token stream.
// It creates and returns a BreakStatement node. If the next token is a semicolon,
// it advances the parser to the next token to consume it.
func (p *Parser) parseBreakStatement() *BreakStatement {
	stmt := &BreakStatement{Token: p.curToken}
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseContinueStatement parses a 'continue' statement from the current token stream.
// It creates and returns a ContinueStatement node. If the next token is a semicolon,
// it advances the parser to the next token before returning the statement.
func (p *Parser) parseContinueStatement() *ContinueStatement {
	stmt := &ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseExpressionStatement parses an expression statement from the current token stream.
// It creates an ExpressionStatement node, parses the contained expression with the lowest precedence,
// and advances the token if a semicolon is present after the expression.
// Returns the constructed ExpressionStatement node.
func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Operator precedence
const (
	_ int = iota
	LOWEST
	ASSIGN_PREC // =
	CONDITIONAL // ?:
	LOGOR       // ||
	LOGAND      // &&
	BITOR_PREC  // |
	BITXOR_PREC // ^
	BITAND_PREC // &
	EQUALS      // == !=
	LESSGREATER // > < >= <=
	SHIFT       // << >>
	SUM         // + -
	PRODUCT     // * / %
	PREFIX      // -X !X
	POSTFIX     // X++ X--
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[TokenType]int{
	ASSIGN:    ASSIGN_PREC,
	PLUSEQ:    ASSIGN_PREC,
	MINUSEQ:   ASSIGN_PREC,
	STAREQ:    ASSIGN_PREC,
	SLASHEQ:   ASSIGN_PREC,
	PERCENTEQ: ASSIGN_PREC,
	OR:        LOGOR,
	AND:       LOGAND,
	BITOR:     BITOR_PREC,
	BITXOR:    BITXOR_PREC,
	BITAND:    BITAND_PREC,
	EQ:        EQUALS,
	NEQ:       EQUALS,
	LT:        LESSGREATER,
	GT:        LESSGREATER,
	LTE:       LESSGREATER,
	GTE:       LESSGREATER,
	LSHIFT:    SHIFT,
	RSHIFT:    SHIFT,
	PLUS:      SUM,
	MINUS:     SUM,
	SLASH:     PRODUCT,
	STAR:      PRODUCT,
	PERCENT:   PRODUCT,
	INC:       POSTFIX,
	DEC:       POSTFIX,
	LPAREN:    CALL,
	LBRACKET:  INDEX,
	QUESTION:  CONDITIONAL,
}

// peekPrecedence returns the precedence of the next token (peekToken).
// If the token type is not found in the precedences map, it returns LOWEST.
// This is used to determine the order of operations during parsing.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence returns the precedence value of the current token.
// If the current token's type is not found in the precedences map,
// it returns the lowest precedence level (LOWEST).
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// parseExpression parses an expression starting from the current token with the given precedence.
// It first parses prefix expressions (identifiers, literals, prefix operators, grouped expressions),
// then parses infix expressions (binary operators, assignments, postfix operators, function calls,
// array accesses, and conditional expressions) as long as the next token has higher precedence.
// Returns the parsed Expression node or nil if parsing fails.
func (p *Parser) parseExpression(precedence int) Expression {
	// Prefix
	var leftExp Expression

	switch p.curToken.Type {
	case IDENT:
		leftExp = &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case INT:
		val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		leftExp = &IntegerLiteral{Token: p.curToken, Value: val}
	case FLOAT:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		leftExp = &FloatLiteral{Token: p.curToken, Value: val}
	case STRING:
		leftExp = &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	case CHAR:
		var val byte
		if len(p.curToken.Literal) > 0 {
			val = p.curToken.Literal[0]
		}
		leftExp = &CharLiteral{Token: p.curToken, Value: val}
	case MINUS, NOT, BITNOT, INC, DEC, STAR, BITAND:
		leftExp = p.parsePrefixExpression()
	case LPAREN:
		p.nextToken()
		leftExp = p.parseExpression(LOWEST)
		if !p.expectPeek(RPAREN) {
			return nil
		}
	default:
		return nil
	}

	// Infix
	for !p.peekTokenIs(SEMICOLON) && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case PLUS, MINUS, STAR, SLASH, PERCENT,
			EQ, NEQ, LT, GT, LTE, GTE,
			AND, OR, BITAND, BITOR, BITXOR,
			LSHIFT, RSHIFT:
			p.nextToken()
			leftExp = p.parseInfixExpression(leftExp)
		case ASSIGN, PLUSEQ, MINUSEQ, STAREQ, SLASHEQ, PERCENTEQ:
			p.nextToken()
			leftExp = p.parseAssignmentExpression(leftExp)
		case INC, DEC:
			p.nextToken()
			leftExp = &PostfixExpression{
				Token:    p.curToken,
				Left:     leftExp,
				Operator: p.curToken.Literal,
			}
		case LPAREN:
			p.nextToken()
			leftExp = p.parseCallExpression(leftExp)
		case LBRACKET:
			p.nextToken()
			leftExp = p.parseArrayExpression(leftExp)
		case QUESTION:
			p.nextToken()
			leftExp = p.parseConditionalExpression(leftExp)
		default:
			return leftExp
		}
	}

	return leftExp
}

// parsePrefixExpression parses a prefix expression from the current token stream.
// It constructs a PrefixExpression node using the current token as the operator,
// advances to the next token, and recursively parses the right-hand side expression
// with PREFIX precedence. Returns the constructed PrefixExpression node.
func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parses an infix expression with the given left-side expression.
// It creates an InfixExpression node, sets its operator and left expression, then parses
// the right-side expression based on the current operator precedence.
// Returns the constructed InfixExpression node.
func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseAssignmentExpression parses an assignment expression starting from the given left-hand side expression.
// It constructs an AssignmentExpression node with the current token as the assignment operator,
// advances the parser to the next token, and parses the right-hand side expression with assignment precedence.
// Returns the constructed AssignmentExpression as an Expression.
func (p *Parser) parseAssignmentExpression(left Expression) Expression {
	expression := &AssignmentExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(ASSIGN_PREC)

	return expression
}

// parseCallExpression parses a function call expression starting from the current token.
// It takes the function expression being called as an argument and returns a CallExpression
// node containing the function and its argument list.
func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(RPAREN)
	return exp
}

// parseArrayExpression parses an array access expression, such as arr[index].
// It takes the left-hand side expression (typically the array identifier) and
// parses the index expression inside the brackets. If the closing bracket is
// not found, it returns nil. Otherwise, it returns an ArrayExpression node
// representing the parsed array access.
func (p *Parser) parseArrayExpression(left Expression) Expression {
	exp := &ArrayExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(RBRACKET) {
		return nil
	}

	return exp
}

// parseConditionalExpression parses a conditional (ternary) expression of the form
// "condition ? consequence : alternative". It takes the already-parsed condition
// expression as input, then parses the consequence and alternative expressions,
// constructing and returning a ConditionalExpression AST node. If the expected colon
// (:) is not found after the consequence, it returns nil.
func (p *Parser) parseConditionalExpression(condition Expression) Expression {
	exp := &ConditionalExpression{
		Token:     p.curToken,
		Condition: condition,
	}

	p.nextToken()
	exp.Consequence = p.parseExpression(LOWEST)

	if !p.expectPeek(COLON) {
		return nil
	}

	p.nextToken()
	exp.Alternative = p.parseExpression(CONDITIONAL)

	return exp
}

// parseExpressionList parses a comma-separated list of expressions until the specified end token is encountered.
// It returns a slice of Expression containing the parsed expressions. If the end token is not found as expected,
// it returns nil.
func (p *Parser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}
