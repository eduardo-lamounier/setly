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

	Line int
	Col  int
	Loc  int
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

func (t Token) String() string {
	kind := tokenKindsToString[t.Kind]
	switch t.Kind {
	case TokenID, TokenString, TokenNumber:
		kind += fmt.Sprintf("(%v)", t.Literal)
	}

	return fmt.Sprintf("[%d:%d - %d, %s, \"%s\"]",
		t.Line, t.Col, t.Loc, kind, t.Lexeme)
}
