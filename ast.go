package cint

// Node represents a node in the abstract syntax tree (AST).
// All AST nodes must implement the TokenLiteral and String methods,
// which return the literal value of the token associated with the node
// and a string representation of the node, respectively.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a node in the abstract syntax tree (AST) that is classified as a statement.
// All statement types must implement the Node interface and the statementNode marker method.
type Statement interface {
	Node
	statementNode()
}

// Expression represents a node in the abstract syntax tree (AST) that is an expression.
// It embeds the Node interface and requires the implementation of the expressionNode method
// to distinguish expression nodes from other node types.
type Expression interface {
	Node
	expressionNode()
}

// Program represents the root node of the AST, containing a slice of Statement nodes
// that make up the entire program.
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the literal value of the first statement's token in the program.
// If the program contains no statements, it returns an empty string.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String returns the string representation of the Program by concatenating
// the string representations of all its statements.
func (p *Program) String() string {
	out := ""
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// FunctionDecl represents a function declaration in the abstract syntax tree (AST).
// It contains the function's name token, return type, name, list of parameters, and the function body.
type FunctionDecl struct {
	Token      Token // the function name token
	ReturnType string
	Name       string
	Parameters []*Parameter
	Body       *BlockStatement
}

func (fd *FunctionDecl) statementNode()       {}
func (fd *FunctionDecl) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDecl) String() string       { return fd.ReturnType + " " + fd.Name + "(...)" }

// Parameter represents a function or method parameter with its type and name.
type Parameter struct {
	Type string
	Name string
}

// VarDecl represents a variable declaration in the abstract syntax tree (AST).
// It contains the token associated with the declaration, the variable's type,
// its name, and an optional initial value expression.
type VarDecl struct {
	Token Token
	Type  string
	Name  string
	Value Expression
}

func (vd *VarDecl) statementNode()       {}
func (vd *VarDecl) TokenLiteral() string { return vd.Token.Literal }
func (vd *VarDecl) String() string       { return vd.Type + " " + vd.Name }

// BlockStatement represents a block of statements enclosed by a pair of braces.
// It contains the opening token and a slice of statements that are executed sequentially.
type BlockStatement struct {
	Token      Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string       { return "{...}" }

// ReturnStatement represents a return statement in the abstract syntax tree (AST).
// It holds the 'return' token and the expression to be returned.
type ReturnStatement struct {
	Token       Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string       { return "return" }

// ExpressionStatement represents a statement consisting of a single expression.
// It holds the initial token of the statement and the associated expression node.
type ExpressionStatement struct {
	Token      Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IfStatement represents an 'if' statement in the abstract syntax tree (AST).
// It contains the token for the 'if' keyword, the condition expression,
// the consequence block to execute if the condition is true, and an optional
// alternative block for the 'else' case.
type IfStatement struct {
	Token       Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string       { return "if" }


// WhileStatement represents a 'while' loop statement in the AST.
// It contains the token for the 'while' keyword, the loop condition expression,
// and the body of the loop as a block statement.
type WhileStatement struct {
	Token     Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string       { return "while" }


// ForStatement represents a 'for' loop construct in the AST.
// It contains the initial statement (Init), loop condition (Condition),
// post-iteration expression (Post), and the loop body (Body).
// The Token field holds the token associated with the 'for' keyword.
type ForStatement struct {
	Token     Token
	Init      Statement
	Condition Expression
	Post      Expression
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string       { return "for" }


// BreakStatement represents a 'break' statement in the abstract syntax tree (AST).
// It contains the token associated with the 'break' keyword.
type BreakStatement struct {
	Token Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break" }


// ContinueStatement represents a 'continue' statement in the abstract syntax tree (AST).
// It holds the token associated with the 'continue' keyword.
type ContinueStatement struct {
	Token Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue" }


// Identifier represents an identifier node in the abstract syntax tree (AST).
// It holds the token associated with the identifier and its string value.
type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }


// IntegerLiteral represents an integer constant in the abstract syntax tree (AST).
// It holds the token associated with the literal and its integer value.
type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }


// FloatLiteral represents a floating-point literal in the abstract syntax tree (AST).
// It contains the token associated with the literal and its float64 value.
type FloatLiteral struct {
	Token Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }


// StringLiteral represents a string literal in the abstract syntax tree (AST).
// It contains the token associated with the literal and its string value.
type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Value }


// CharLiteral represents a character literal in the abstract syntax tree (AST).
// It contains the token associated with the literal and its byte value.
type CharLiteral struct {
	Token Token
	Value byte
}

func (cl *CharLiteral) expressionNode()      {}
func (cl *CharLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *CharLiteral) String() string       { return cl.Token.Literal }


// PrefixExpression represents an expression with a prefix operator (such as ! or -) applied to a single operand.
// It contains the operator token, the operator string, and the right-hand side expression.
type PrefixExpression struct {
	Token    Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string       { return "(" + pe.Operator + pe.Right.String() + ")" }


// PostfixExpression represents an expression with a postfix operator (e.g., x++ or x--).
// It contains the token for the operator, the left-hand side expression, and the operator string.
type PostfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
}

func (pe *PostfixExpression) expressionNode()      {}
func (pe *PostfixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PostfixExpression) String() string       { return "(" + pe.Left.String() + pe.Operator + ")" }

// InfixExpression represents an infix operation in the abstract syntax tree (AST).
// It contains the operator token, the left and right expressions, and the operator string itself.
type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}


// CallExpression represents a function call expression in the AST.
// It contains the token for the call, the function being called, and the list of argument expressions.
type CallExpression struct {
	Token     Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string       { return ce.Function.String() + "(...)" }


// AssignmentExpression represents an assignment operation in the abstract syntax tree (AST).
// It contains the assignment token, the left-hand side expression, the assignment operator,
// and the right-hand side expression.
type AssignmentExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ae *AssignmentExpression) expressionNode()      {}
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	return ae.Left.String() + " " + ae.Operator + " " + ae.Right.String()
}


// ArrayExpression represents an array access expression, containing the token for the array access,
// the expression for the array being accessed (Left), and the expression for the index (Index).
type ArrayExpression struct {
	Token Token
	Left  Expression
	Index Expression
}

func (ae *ArrayExpression) expressionNode()      {}
func (ae *ArrayExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *ArrayExpression) String() string       { return ae.Left.String() + "[" + ae.Index.String() + "]" }


// ConditionalExpression represents a conditional (ternary) expression in the AST,
// consisting of a condition, a consequence (expression if the condition is true),
// and an alternative (expression if the condition is false).
type ConditionalExpression struct {
	Token       Token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (ce *ConditionalExpression) expressionNode()      {}
func (ce *ConditionalExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *ConditionalExpression) String() string       { return "(...? ... : ...)" }
