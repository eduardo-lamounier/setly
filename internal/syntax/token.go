// Package syntax Defines the information related to a token that is stored.
// Also defines the different kinds of tokens.
package syntax

import "fmt"

type TokenKind int

const (
	TokenParenL TokenKind = iota
	TokenParenR
	TokenBraceL
	TokenBraceR

	TokenComma
	TokenDot
	TokenColon
	TokenSemicolon

	TokenNewline

	TokenMinus
	TokenPlus
	TokenStar
	TokenSlash
	TokenSlashSlash

	TokenColonEqual
	TokenEqual
	TokenTildeEqual
	TokenGreater
	TokenLess
	TokenGreaterEqual
	TokenLessEqual

	TokenID
	TokenString
	TokenNumber

	TokenFalse
	TokenTrue
	TokenNil
	TokenAnd
	TokenOr
	TokenNot
	TokenIf
	TokenElse
	TokenElif
	TokenWhile
	TokenFor
	TokenPrint
	TokenPrintln
	TokenFn
	TokenReturn
	TokenLet

	TokenEOF
)

type Token struct {
	Kind    TokenKind
	Lexeme  string
	Literal any

	Location
}

type Location struct {
	pos  int
	line int
	col  int
}

type locatable interface {
	Pos() int
	Line() int
	Col() int
}

var KeywordForLexeme = map[string]TokenKind{
	"false":   TokenFalse,
	"true":    TokenTrue,
	"nil":     TokenNil,
	"and":     TokenAnd,
	"or":      TokenOr,
	"not":     TokenNot,
	"if":      TokenIf,
	"else":    TokenElse,
	"elif":    TokenElif,
	"while":   TokenWhile,
	"for":     TokenFor,
	"print":   TokenPrint,
	"println": TokenPrintln,
	"fn":      TokenFn,
	"return":  TokenReturn,
	"let":     TokenLet,
}

var lexemeForTokenKind = map[TokenKind]string{
	TokenParenL: "(",
	TokenParenR: ")",
	TokenBraceL: "{",
	TokenBraceR: "}",

	TokenComma:     ",",
	TokenDot:       ".",
	TokenColon:     ":",
	TokenSemicolon: ";",

	TokenNewline: "\n",

	TokenMinus:      "-",
	TokenPlus:       "+",
	TokenStar:       "*",
	TokenSlash:      "/",
	TokenSlashSlash: "//",

	TokenColonEqual:   ":=",
	TokenEqual:        "=",
	TokenTildeEqual:   "~=",
	TokenGreater:      ">",
	TokenLess:         "<",
	TokenGreaterEqual: ">=",
	TokenLessEqual:    "<=",

	// Literals do not have a specific lexeme

	TokenFalse:   "false",
	TokenTrue:    "true",
	TokenNil:     "nil",
	TokenAnd:     "and",
	TokenOr:      "or",
	TokenNot:     "not",
	TokenIf:      "if",
	TokenElse:    "else",
	TokenElif:    "elif",
	TokenWhile:   "while",
	TokenFor:     "for",
	TokenPrint:   "print",
	TokenPrintln: "println",
	TokenFn:      "fn",
	TokenReturn:  "return",
	TokenLet:     "let",

	TokenEOF: "",
}

var tokenKindsToString = map[TokenKind]string{
	TokenParenL: "LEFT_PAREN",
	TokenParenR: "RIGHT_PAREN",
	TokenBraceL: "LEFT_BRACE",
	TokenBraceR: "RIGHT_BRACE",

	TokenComma:     "COMMA",
	TokenDot:       "DOT",
	TokenColon:     "COLON",
	TokenSemicolon: "SEMICOLON",

	TokenNewline: "NEWLINE",

	TokenMinus:      "MINUS",
	TokenPlus:       "PLUS",
	TokenStar:       "STAR",
	TokenSlash:      "SLASH",
	TokenSlashSlash: "SLASH_SLASH",

	TokenColonEqual:   "COLON_EQUAL",
	TokenEqual:        "EQUAL",
	TokenTildeEqual:   "TILDE_EQUAL",
	TokenGreater:      "GREATER",
	TokenLess:         "LESS",
	TokenGreaterEqual: "GREATER_EQUAL",
	TokenLessEqual:    "LESS_EQUAL",

	TokenID:     "ID",
	TokenString: "STRING",
	TokenNumber: "NUMBER",

	TokenFalse:   "FALSE",
	TokenTrue:    "TRUE",
	TokenNil:     "NIL",
	TokenAnd:     "AND",
	TokenOr:      "OR",
	TokenNot:     "NOT",
	TokenIf:      "IF",
	TokenElse:    "ELSE",
	TokenElif:    "ELIF",
	TokenWhile:   "WHILE",
	TokenFor:     "FOR",
	TokenPrint:   "PRINT",
	TokenPrintln: "PRINTLN",
	TokenFn:      "FN",
	TokenReturn:  "RETURN",
	TokenLet:     "LET",

	TokenEOF: "EOF",
}

func (l Location) Pos() int {
	return l.pos
}

func (l Location) Line() int {
	return l.line
}

func (l Location) Col() int {
	return l.col
}

func NewToken(kind TokenKind, lexeme string, pos int, line int, col int) Token {
	return NewLiteralToken(kind, lexeme, nil, pos, line, col)
}

func NewLiteralToken(kind TokenKind, lexeme string, literal any, pos int, line int, col int) Token {
	return Token{
		Kind:    kind,
		Lexeme:  lexeme,
		Literal: literal,

		Location: Location{pos, line, col},
	}
}

func (tk TokenKind) GetLexeme() string {
	if tk == TokenID || tk == TokenString || tk == TokenNumber {
		panic("Tried getting the lexeme of a literal token kind.")
	}

	return lexemeForTokenKind[tk]
}

func (t Token) String() string {
	kind := tokenKindsToString[t.Kind]
	switch t.Kind {
	case TokenID, TokenString, TokenNumber:
		kind += fmt.Sprintf("(%v)", t.Literal)
	}

	return fmt.Sprintf("[%d:%d - %d, %s, \"%s\"]",
		t.Line(), t.Col(), t.Pos(), kind, t.Lexeme)
}
