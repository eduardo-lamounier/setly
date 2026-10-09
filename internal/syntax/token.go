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
	TokenStr
	TokenNum

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
