package cint

import (
	"unicode"
)


// Lexer represents a lexical analyzer for tokenizing input strings.
// It maintains the input string, current position, reading position,
// current character, and tracking information for the current line and column.
type Lexer struct {
	input        string
	position     int  // current position in input
	readPosition int  // current reading position
	ch           byte // current char
	line         int
	column       int
}


// NewLexer creates and initializes a new Lexer instance for the given input string.
// It sets the starting line and column, reads the first character, and returns a pointer to the Lexer.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

// readChar advances the lexer by one character, updating the current character (l.ch),
// position, readPosition, line, and column counters. If the end of input is reached,
// l.ch is set to 0. Handles line and column tracking for newline characters.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

// peekChar returns the next character in the input without advancing the lexer position.
// If the end of the input is reached, it returns 0.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}


// NextToken scans the input and returns the next Token from the input stream.
// It skips whitespace and comments, then determines the type of token based on the current character.
// The function handles single and multi-character operators, delimiters, string and character literals,
// identifiers, numbers, and special tokens such as EOF and ILLEGAL. The returned Token includes
// the token type, literal value, and the line and column where the token was found.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()
	l.skipComments()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: ASSIGN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: INC, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PLUSEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: PLUS, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: DEC, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MINUSEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: MINUS, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: STAREQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: STAR, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: SLASHEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: SLASH, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PERCENTEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: PERCENT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: NOT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				lit := string(ch) + string(l.ch)
				l.readChar()
				tok = Token{Type: LSHIFTEQ, Literal: lit + string(l.ch), Line: tok.Line, Column: tok.Column}
			} else {
				tok = Token{Type: LSHIFT, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
			}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			if l.peekChar() == '=' {
				lit := string(ch) + string(l.ch)
				l.readChar()
				tok = Token{Type: RSHIFTEQ, Literal: lit + string(l.ch), Line: tok.Line, Column: tok.Column}
			} else {
				tok = Token{Type: RSHIFT, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
			}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ANDEQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: BITAND, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OREQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: BITOR, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '^':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: XOREQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: BITXOR, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	case '~':
		tok = Token{Type: BITNOT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '{':
		tok = Token{Type: LBRACE, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '}':
		tok = Token{Type: RBRACE, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '[':
		tok = Token{Type: LBRACKET, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case ']':
		tok = Token{Type: RBRACKET, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '.':
		tok = Token{Type: DOT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '?':
		tok = Token{Type: QUESTION, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case ':':
		tok = Token{Type: COLON, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	case '\'':
		tok.Type = CHAR
		tok.Literal = l.readCharLiteral()
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace advances the lexer position past any whitespace characters,
// including spaces, tabs, newlines, and carriage returns.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipComments skips over single-line (//) and multi-line (/* ... */) comments in the input.
// It advances the lexer position past any comments and any subsequent whitespace.
// For multi-line comments, it also handles consecutive comments by recursively calling itself.
func (l *Lexer) skipComments() {
	if l.ch == '/' {
		if l.peekChar() == '/' {
			// Single line comment
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			l.skipWhitespace()
		} else if l.peekChar() == '*' {
			// Multi-line comment
			l.readChar()
			l.readChar()
			for {
				if l.ch == 0 {
					break
				}
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar()
					l.readChar()
					break
				}
				l.readChar()
			}
			l.skipWhitespace()
			l.skipComments() // Handle consecutive comments
		}
	}
}

// readIdentifier reads an identifier from the current position in the input.
// It advances the lexer until a non-letter and non-digit character is encountered,
// then returns the substring representing the identifier.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a numeric literal from the input and determines its type.
// It supports integer, floating-point, and scientific notation formats, as well as
// optional suffixes such as L, U, and F (case-insensitive) commonly used in C-like languages.
// The function returns the string representation of the number and its corresponding TokenType.
func (l *Lexer) readNumber() (string, TokenType) {
	position := l.position
	tokType := INT

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' {
		tokType = FLOAT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	if l.ch == 'e' || l.ch == 'E' {
		tokType = FLOAT
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Skip suffixes like L, U, F
	if l.ch == 'L' || l.ch == 'l' || l.ch == 'U' || l.ch == 'u' || l.ch == 'F' || l.ch == 'f' {
		l.readChar()
		if l.ch == 'L' || l.ch == 'l' || l.ch == 'U' || l.ch == 'u' {
			l.readChar()
		}
	}

	return l.input[position:l.position], tokType
}

// readString reads a string literal from the input, handling escape sequences.
// It assumes the opening quote has already been encountered and advances until
// it finds the closing quote or the end of input. The function returns the
// string content without the surrounding quotes.
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

// readCharLiteral reads a character literal from the input, handling escape sequences.
// It assumes the current position is at the opening single quote and reads until the closing single quote or end of input.
// Returns the string representing the character literal (excluding the surrounding single quotes).
func (l *Lexer) readCharLiteral() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

// isLetter checks if the given byte represents a Unicode letter or an underscore ('_').
// It returns true if the character is a letter or underscore, and false otherwise.
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// isDigit returns true if the given byte represents an ASCII digit ('0' to '9'), and false otherwise.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
