// Package syntax Defines the information related to a token that is stored.
// Also defines the different kinds of tokens.
package syntax

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
