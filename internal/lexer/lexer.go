// Package lexer TODO: Defines and implements functions for converting the
// source to a sequence of tokens
package lexer

import (
	"fmt"
	"slices"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/eduardo-lamounier/setly/internal/diagnostics"
	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type lexer struct {
	Source string
	Pos    int
	Start  int
	Line   int
	Col    int

	HadError    bool
	Diagnostics []diagnostics.Diagnostic
	Tokens      []syntax.Token
}

func (l lexer) atEnd() bool {
	return l.Pos >= len(l.Source)
}

func (l *lexer) advanceBy(n int) {
	for range n {
		if l.atEnd() {
			return
		}

		r, size := utf8.DecodeRuneInString(l.Source[l.Pos:])

		if r == '\n' {
			l.Line++
			l.Col = 0
		} else {
			l.Col++
		}

		l.Pos += size
	}
}

func (l *lexer) advance() { l.advanceBy(1) }

func (l *lexer) finish() {
	l.Pos = len(l.Source)
}

func (l lexer) peekNext() rune {
	if l.Pos+1 >= len(l.Source) {
		return 0
	}

	r, _ := utf8.DecodeRuneInString(l.Source[l.Pos+1:])
	return r
}

func (l lexer) peek() rune {
	if l.atEnd() {
		return 0
	}

	r, _ := utf8.DecodeRuneInString(l.Source[l.Pos:])
	return r
}

func (l lexer) peekSub(n int) string {
	return l.Source[l.Pos:min(len(l.Source), l.Pos+n)]
}

func (l *lexer) matchSub(n int, expected ...string) bool {
	if n < 0 {
		n = len(expected)
	}

	if l.Pos+n >= len(l.Source) {
		return false
	}

	if !slices.Contains(expected, l.Source[l.Pos:l.Pos+n]) {
		return false
	}

	l.advanceBy(n)
	return true
}

func (l *lexer) match(expected ...byte) bool {
	if l.atEnd() {
		return false
	}

	if !slices.Contains(expected, l.Source[l.Pos]) {
		return false
	}

	l.advance()
	return true
}

func (l *lexer) lexemeBegins() {
	l.Start = l.Pos
}

func (l *lexer) report(message string) {
	l.reportAt(l.Line, l.Col, l.Pos, message)
}

func (l *lexer) reportAt(line int, col int, loc int, message string) {
	l.HadError = true
	l.Diagnostics = append(l.Diagnostics, diagnostics.Diagnostic{
		Source:  l.Source,
		Line:    line,
		Col:     col,
		Loc:     loc,
		Message: message,
	})
}

func (l *lexer) addToken(kind syntax.TokenKind) {
	l.addLiteralToken(kind, nil)
}

func (l *lexer) addLiteralToken(
	kind syntax.TokenKind, literal any,
) {
	line := l.Line
	col := l.Col

	for i := l.Pos - 1; i > l.Start; i-- {
		if l.Source[i] == '\n' {
			line--
			col = 0
		} else {
			col--
		}
	}

	l.Tokens = append(l.Tokens, syntax.Token{
		Kind:    kind,
		Lexeme:  l.Source[l.Start:l.Pos],
		Literal: literal,

		Line: line,
		Col:  col,
		Loc:  l.Start,
	})
}

func (l *lexer) scanToken() {
	l.lexemeBegins()
	chr := l.peek()
	l.advance()

	switch chr {
	case '@':
		// Line comment:
		if !l.match('@') {
			for ; !l.atEnd(); l.advance() {
				if l.peek() == '\n' {
					l.lexemeBegins()
					l.advance()
					l.addToken(syntax.TokenNewline)
					break
				}
			}
			break
		}

		line, col := l.Line, l.Col

		// Block comment:
		newline := false
		for ; !l.atEnd() && l.peekSub(2) != "@@"; l.advance() {
			if l.peek() == '\n' && !newline {
				newline = true
				l.lexemeBegins()
				l.advance()
				l.addToken(syntax.TokenNewline)
			}
		}

		if l.atEnd() {
			l.reportAt(line, col, l.Start, "Comment block wasn't closed.")
			break
		}

		l.advanceBy(2) // Consumes the "@@"

	case '(':
		l.addToken(syntax.TokenParenL)
	case ')':
		l.addToken(syntax.TokenParenR)
	case '{':
		l.addToken(syntax.TokenBraceL)
	case '}':
		l.addToken(syntax.TokenBraceR)
	case ',':
		l.addToken(syntax.TokenComma)
	case '.':
		l.addToken(syntax.TokenDot)
	case ':':
		if l.match('=') {
			l.addToken(syntax.TokenColonEqual)
		} else {
			l.addToken(syntax.TokenColon)
		}
	case ';':
		l.addToken(syntax.TokenSemicolon)
	case '+':
		l.addToken(syntax.TokenPlus)
	case '-':
		l.addToken(syntax.TokenMinus)
	case '*':
		l.addToken(syntax.TokenStar)
	case '/':
		if l.match('/') {
			l.addToken(syntax.TokenSlashSlash)
		} else {
			l.addToken(syntax.TokenSlash)
		}
	case '=':
		l.addToken(syntax.TokenEqual)
	case '~':
		if l.match('=') {
			l.addToken(syntax.TokenTildeEqual)
		} else {
			l.reportAt(l.Line, l.Col-1, l.Pos-1, "Unexpected chracter \"~\"")
		}
	case '<':
		if l.match('=') {
			l.addToken(syntax.TokenLessEqual)
		} else {
			l.addToken(syntax.TokenLess)
		}
	case '>':
		if l.match('=') {
			l.addToken(syntax.TokenGreaterEqual)
		} else {
			l.addToken(syntax.TokenGreater)
		}
	case ' ', '\r', '\t':
	case '\n':
		l.addToken(syntax.TokenNewline)

	case '"': // String literal
		line, col := l.Line, l.Col

		for ; l.peek() != '"' && l.peek() != '\n' && !l.atEnd(); l.advance() {
		}

		if l.peek() == '\n' || l.atEnd() {
			l.reportAt(line, col, l.Start, "Unterminated string.")
			// Skips the following lines as too many
			// unreadable errors may occur:
			l.finish()
			break
		}

		l.advance() // Consumes the last "
		s := l.Source[l.Start+1 : l.Pos-1]
		l.addLiteralToken(syntax.TokenString, s)

	default:
		// Number literal:
		if unicode.IsDigit(chr) {
			for ; unicode.IsDigit(l.peek()); l.advance() {
			}

			if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
				l.advance() // Consumes the .

				for ; unicode.IsDigit(l.peek()); l.advance() {
				}
			}

			n, err := strconv.ParseFloat(l.Source[l.Start:l.Pos], 64)
			if err != nil {
				panic("Could not convert number literal to a float.")
			}

			l.addLiteralToken(syntax.TokenNumber, n)
			break
		}

		// Identifier:
		if unicode.IsLetter(chr) {
			for ; unicode.IsLetter(l.peek()) || unicode.IsNumber(l.peek()); l.advance() {
			}

			s := l.Source[l.Start:l.Pos]
			if tk, isKeyword := syntax.KeywordForLexeme[s]; isKeyword {
				l.addToken(tk)
			} else {
				l.addLiteralToken(syntax.TokenID, s)
			}

			break
		}

		l.reportAt(l.Line, l.Col-1, l.Pos-1,
			fmt.Sprintf("Unexpected character \"%c\"", chr))
	}
}

func (l *lexer) insertSemicolons() {
	for i, tok := range l.Tokens {
		if tok.Kind == syntax.TokenNewline && i > 0 {
			switch l.Tokens[i-1].Kind {
			case syntax.TokenParenR,
				syntax.TokenBraceR,
				syntax.TokenString,
				syntax.TokenNumber,
				syntax.TokenFalse,
				syntax.TokenTrue,
				syntax.TokenNil,
				syntax.TokenID,
				syntax.TokenReturn:

				l.Tokens[i].Kind = syntax.TokenSemicolon
			}
		}
	}
}

func Scan(source string) ([]syntax.Token, bool, []diagnostics.Diagnostic) {
	l := lexer{
		Source: source,
		Pos:    0,
		Start:  0,
		Line:   1,
		Col:    1,

		Diagnostics: []diagnostics.Diagnostic{},
		Tokens:      []syntax.Token{},
	}

	for ; !l.atEnd(); l.scanToken() {
	}

	l.lexemeBegins()
	l.addToken(syntax.TokenEOF)

	l.insertSemicolons() // Replaces some newlines with semicolons

	return l.Tokens, l.HadError, l.Diagnostics
}
