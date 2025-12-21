package cint

// TokenType represents the type of token
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	ILLEGAL

	// Identifiers and literals
	IDENT
	INT
	FLOAT
	CHAR
	STRING

	// Keywords
	AUTO
	BREAK
	CASE
	CHAR_KW
	CONST
	CONTINUE
	DEFAULT
	DO
	DOUBLE
	ELSE
	ENUM
	EXTERN
	FLOAT_KW
	FOR
	GOTO
	IF
	INT_KW
	LONG
	REGISTER
	RETURN
	SHORT
	SIGNED
	SIZEOF
	STATIC
	STRUCT
	SWITCH
	TYPEDEF
	UNION
	UNSIGNED
	VOID
	VOLATILE
	WHILE

	// Operators
	PLUS      // +
	MINUS     // -
	STAR      // *
	SLASH     // /
	PERCENT   // %
	ASSIGN    // =
	EQ        // ==
	NEQ       // !=
	LT        // <
	GT        // >
	LTE       // <=
	GTE       // >=
	AND       // &&
	OR        // ||
	NOT       // !
	BITAND    // &
	BITOR     // |
	BITXOR    // ^
	BITNOT    // ~
	LSHIFT    // <<
	RSHIFT    // >>
	INC       // ++
	DEC       // --
	PLUSEQ    // +=
	MINUSEQ   // -=
	STAREQ    // *=
	SLASHEQ   // /=
	PERCENTEQ // %=
	ANDEQ     // &=
	OREQ      // |=
	XOREQ     // ^=
	LSHIFTEQ  // <<=
	RSHIFTEQ  // >>=

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	SEMICOLON // ;
	COMMA     // ,
	DOT       // .
	ARROW     // ->
	QUESTION  // ?
	COLON     // :
)

var keywords = map[string]TokenType{
	"auto":     AUTO,
	"break":    BREAK,
	"case":     CASE,
	"char":     CHAR_KW,
	"const":    CONST,
	"continue": CONTINUE,
	"default":  DEFAULT,
	"do":       DO,
	"double":   DOUBLE,
	"else":     ELSE,
	"enum":     ENUM,
	"extern":   EXTERN,
	"float":    FLOAT_KW,
	"for":      FOR,
	"goto":     GOTO,
	"if":       IF,
	"int":      INT_KW,
	"long":     LONG,
	"register": REGISTER,
	"return":   RETURN,
	"short":    SHORT,
	"signed":   SIGNED,
	"sizeof":   SIZEOF,
	"static":   STATIC,
	"struct":   STRUCT,
	"switch":   SWITCH,
	"typedef":  TYPEDEF,
	"union":    UNION,
	"unsigned": UNSIGNED,
	"void":     VOID,
	"volatile": VOLATILE,
	"while":    WHILE,
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// LookupIdent returns the token type for an identifier
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
