package cint

import (
	"fmt"
	"strconv"
)

// Parser parses C code into an AST
type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

// NewParser creates a new parser
func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %v (type %d), got %v (type %d) '%s' instead at line %d",
		t, t, p.peekToken.Type, p.peekToken.Type, p.peekToken.Literal, p.peekToken.Line)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses the entire program
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

func (p *Parser) isTypeKeyword(t TokenType) bool {
	return t == INT_KW || t == CHAR_KW || t == FLOAT_KW || t == DOUBLE ||
		t == VOID || t == LONG || t == SHORT || t == UNSIGNED || t == SIGNED
}

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

func (p *Parser) parseBreakStatement() *BreakStatement {
	stmt := &BreakStatement{Token: p.curToken}
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseContinueStatement() *ContinueStatement {
	stmt := &ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

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

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

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

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

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

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(RPAREN)
	return exp
}

func (p *Parser) parseArrayExpression(left Expression) Expression {
	exp := &ArrayExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(RBRACKET) {
		return nil
	}

	return exp
}

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
