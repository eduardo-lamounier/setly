package lexer

import (
	"reflect"
	"testing"

	"github.com/eduardo-lamounier/setly/internal/syntax"
)

type wantToken struct {
	kind    syntax.TokenKind
	lexeme  string
	literal any
}

func token(
	kind syntax.TokenKind,
	lexeme string,
	literal any,
) wantToken {
	return wantToken{kind, lexeme, literal}
}

func sameTokens(got []syntax.Token, want []wantToken) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i].Kind != want[i].kind || got[i].Lexeme != want[i].lexeme ||
			!reflect.DeepEqual(got[i].LiteralValue, want[i].literal) {
			return false
		}
	}

	return true
}

func TestScan(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    []wantToken
		wantErr bool
	}{
		{
			name:   "empty source includes EOF",
			source: "",
			want:   []wantToken{token(syntax.TokenEOF, "", nil)},
		},
		{
			name:   "punctuation and operators",
			source: "() [] {} , . : ; + - -> * / // | = == ~= < <= > >=",
			want: []wantToken{
				token(syntax.TokenParenL, "(", nil),
				token(syntax.TokenParenR, ")", nil),
				token(syntax.TokenBracketL, "[", nil),
				token(syntax.TokenBracketR, "]", nil),
				token(syntax.TokenBraceL, "{", nil),
				token(syntax.TokenBraceR, "}", nil),
				token(syntax.TokenComma, ",", nil),
				token(syntax.TokenDot, ".", nil),
				token(syntax.TokenColon, ":", nil),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenPlus, "+", nil),
				token(syntax.TokenMinus, "-", nil),
				token(syntax.TokenArrowR, "->", nil),
				token(syntax.TokenStar, "*", nil),
				token(syntax.TokenSlash, "/", nil),
				token(syntax.TokenSlashSlash, "//", nil),
				token(syntax.TokenPipe, "|", nil),
				token(syntax.TokenEqual, "=", nil),
				token(syntax.TokenEqualEqual, "==", nil),
				token(syntax.TokenTildeEqual, "~=", nil),
				token(syntax.TokenLess, "<", nil),
				token(syntax.TokenLessEqual, "<=", nil),
				token(syntax.TokenGreater, ">", nil),
				token(syntax.TokenGreaterEqual, ">=", nil),
				token(syntax.TokenEOF, "", nil),
			},
		},
		{
			name:   "keywords and identifiers",
			source: "let x = true and false",
			want: []wantToken{
				token(syntax.TokenLet, "let", nil),
				token(syntax.TokenID, "x", "x"),
				token(syntax.TokenEqual, "=", nil),
				token(syntax.TokenTrue, "true", nil),
				token(syntax.TokenAnd, "and", nil),
				token(syntax.TokenFalse, "false", nil),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenEOF, "", nil),
			},
		},
		{
			name:   "number and string literals",
			source: `42 3.5 "hello"`,
			want: []wantToken{
				token(syntax.TokenNumber, "42", float64(42)),
				token(syntax.TokenNumber, "3.5", float64(3.5)),
				token(syntax.TokenString, `"hello"`, "hello"),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenEOF, "", nil),
			},
		},
		{
			name:   "line comment and whitespace",
			source: "a  @ ignored\nb",
			want: []wantToken{
				token(syntax.TokenID, "a", "a"),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenID, "b", "b"),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenEOF, "", nil),
			},
		},
		{
			name:   "block comment",
			source: "a @@ hidden @@ b",
			want: []wantToken{
				token(syntax.TokenID, "a", "a"),
				token(syntax.TokenID, "b", "b"),
				token(syntax.TokenSemicolon, ";", nil),
				token(syntax.TokenEOF, "", nil),
			},
		},
		{
			name:    "unterminated string",
			source:  `"unfinished`,
			want:    []wantToken{token(syntax.TokenEOF, "", nil)},
			wantErr: true,
		},
		{
			name:    "unterminated block comment",
			source:  "@@ unfinished",
			want:    []wantToken{token(syntax.TokenEOF, "", nil)},
			wantErr: true,
		},
		{
			name:    "unexpected character",
			source:  "~",
			want:    []wantToken{token(syntax.TokenEOF, "", nil)},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr, diagnostics := Scan(test.source)

			if !sameTokens(got, test.want) {
				t.Errorf("Scan(%q) tokens = %#v; want %#v", test.source, got, test.want)
			}
			if gotErr != test.wantErr {
				t.Errorf("Scan(%q) error = %t; want %t", test.source, gotErr, test.wantErr)
			}
			if gotErr && len(diagnostics) == 0 {
				t.Errorf("Scan(%q) returned an error without a diagnostic", test.source)
			}
		})
	}
}
