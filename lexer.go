package prolog

import (
	"fmt"
	"unicode"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenVar
	tokenNumber
	tokenLParen
	tokenRParen
	tokenLBracket
	tokenRBracket
	tokenComma
	tokenDot
	tokenBar
	tokenEquals
	tokenNotUnifiable
	tokenLess
	tokenGreater
	tokenLessEqual
	tokenGreaterEqual
	tokenArithmeticEqual
	tokenArithmeticNotEqual
	tokenPlus
	tokenMinus
	tokenStar
	tokenPower
	tokenSlash
	tokenColonDash
	tokenQuery
)

type token struct {
	typ     tokenType
	literal string
	line    int
	column  int
}

func (t token) String() string {
	switch t.typ {
	case tokenEOF:
		return "EOF"
	case tokenIdent:
		return fmt.Sprintf("identifier %q", t.literal)
	case tokenVar:
		return fmt.Sprintf("variable %q", t.literal)
	case tokenNumber:
		return fmt.Sprintf("number %q", t.literal)
	case tokenLParen:
		return "("
	case tokenRParen:
		return ")"
	case tokenComma:
		return ","
	case tokenLBracket:
		return "["
	case tokenRBracket:
		return "]"
	case tokenDot:
		return "."
	case tokenBar:
		return "|"
	case tokenEquals:
		return "="
	case tokenNotUnifiable:
		return "\\="
	case tokenLess:
		return "<"
	case tokenGreater:
		return ">"
	case tokenLessEqual:
		return "=<"
	case tokenGreaterEqual:
		return ">="
	case tokenArithmeticEqual:
		return "=:="
	case tokenArithmeticNotEqual:
		return "=\\="
	case tokenPlus:
		return "+"
	case tokenMinus:
		return "-"
	case tokenStar:
		return "*"
	case tokenPower:
		return "**"
	case tokenSlash:
		return "/"
	case tokenColonDash:
		return ":-"
	case tokenQuery:
		return "?-"
	default:
		return fmt.Sprintf("unknown token %q", t.literal)
	}
}

type lexer struct {
	input  []rune
	pos    int
	line   int
	column int
}

func lex(src string) ([]token, error) {
	l := lexer{
		input:  []rune(src),
		line:   1,
		column: 1,
	}

	var tokens []token

	for {
		l.skipWhitespaceAndComments()

		startLine := l.line
		startColumn := l.column

		ch := l.peek()

		switch {
		case ch == 0:
			tokens = append(tokens, token{
				typ:    tokenEOF,
				line:   startLine,
				column: startColumn,
			})
			return tokens, nil
		case ch == '(':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenLParen,
				literal: "(",
				line:    startLine,
				column:  startColumn,
			})
		case ch == ')':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenRParen,
				literal: ")",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '[':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenLBracket,
				literal: "[",
				line:    startLine,
				column:  startColumn,
			})
		case ch == ']':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenRBracket,
				literal: "]",
				line:    startLine,
				column:  startColumn,
			})
		case ch == ',':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenComma,
				literal: ",",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '.':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenDot,
				literal: ".",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '|':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenBar,
				literal: "|",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '=':
			l.advance()

			switch l.peek() {
			case '<':
				l.advance()
				tokens = append(tokens, token{
					typ:     tokenLessEqual,
					literal: "=<",
					line:    startLine,
					column:  startColumn,
				})
			case ':':
				l.advance()
				if l.peek() != '=' {
					return nil, fmt.Errorf(
						"%d:%d: expected '=' after '=:'",
						startLine,
						startColumn,
					)
				}
				l.advance()
				tokens = append(tokens, token{
					typ:     tokenArithmeticEqual,
					literal: "=:=",
					line:    startLine,
					column:  startColumn,
				})
			case '\\':
				l.advance()
				if l.peek() != '=' {
					return nil, fmt.Errorf(
						"%d:%d: expected '=' after '=\\\\'",
						startLine,
						startColumn,
					)
				}
				l.advance()
				tokens = append(tokens, token{
					typ:     tokenArithmeticNotEqual,
					literal: "=\\=",
					line:    startLine,
					column:  startColumn,
				})
			default:
				tokens = append(tokens, token{
					typ:     tokenEquals,
					literal: "=",
					line:    startLine,
					column:  startColumn,
				})
			}
		case ch == '\\':
			l.advance()

			if l.peek() != '=' {
				return nil, fmt.Errorf(
					"%d:%d: expected '=' after '\\\\'",
					startLine,
					startColumn,
				)
			}

			l.advance()
			tokens = append(tokens, token{
				typ:     tokenNotUnifiable,
				literal: "\\=",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '<':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenLess,
				literal: "<",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '>':
			l.advance()
			if l.peek() == '=' {
				l.advance()
				tokens = append(tokens, token{
					typ:     tokenGreaterEqual,
					literal: ">=",
					line:    startLine,
					column:  startColumn,
				})
			} else {
				tokens = append(tokens, token{
					typ:     tokenGreater,
					literal: ">",
					line:    startLine,
					column:  startColumn,
				})
			}
		case ch == '+':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenPlus,
				literal: "+",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '-':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenMinus,
				literal: "-",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '*':
			l.advance()
			if l.peek() == '*' {
				l.advance()
				tokens = append(tokens, token{
					typ:     tokenPower,
					literal: "**",
					line:    startLine,
					column:  startColumn,
				})
			} else {
				tokens = append(tokens, token{
					typ:     tokenStar,
					literal: "*",
					line:    startLine,
					column:  startColumn,
				})
			}
		case ch == '/':
			l.advance()
			tokens = append(tokens, token{
				typ:     tokenSlash,
				literal: "/",
				line:    startLine,
				column:  startColumn,
			})
		case ch == ':':
			l.advance()

			if l.peek() != '-' {
				return nil, fmt.Errorf(
					"%d:%d: expected '-' after ':'",
					startLine,
					startColumn,
				)
			}

			l.advance()

			tokens = append(tokens, token{
				typ:     tokenColonDash,
				literal: ":-",
				line:    startLine,
				column:  startColumn,
			})
		case ch == '?':
			l.advance()

			if l.peek() != '-' {
				return nil, fmt.Errorf(
					"%d:%d: expected '-' after '?'",
					startLine,
					startColumn,
				)
			}

			l.advance()

			tokens = append(tokens, token{
				typ:     tokenQuery,
				literal: "?-",
				line:    startLine,
				column:  startColumn,
			})
		case unicode.IsDigit(ch):
			literal := l.readNumber()

			tokens = append(tokens, token{
				typ:     tokenNumber,
				literal: literal,
				line:    startLine,
				column:  startColumn,
			})
		case isIdentifierStart(ch):
			literal := l.readIdentifier()

			tokType := tokenIdent
			first := []rune(literal)[0]

			if unicode.IsUpper(first) || first == '_' {
				tokType = tokenVar
			}

			tokens = append(tokens, token{
				typ:     tokType,
				literal: literal,
				line:    startLine,
				column:  startColumn,
			})
		default:
			return nil, fmt.Errorf(
				"%d:%d: unexpected character %q",
				startLine,
				startColumn,
				ch,
			)
		}
	}
}

func (l *lexer) skipWhitespaceAndComments() {
	for {
		for unicode.IsSpace(l.peek()) {
			l.advance()
		}

		// Prolog-style line comment
		if l.peek() == '%' {
			for l.peek() != 0 && l.peek() != '\n' {
				l.advance()
			}
			continue
		}

		return
	}
}

func (l *lexer) readIdentifier() string {
	start := l.pos

	for isIdentifierPart(l.peek()) {
		l.advance()
	}

	return string(l.input[start:l.pos])
}

func (l *lexer) readNumber() string {
	start := l.pos

	for unicode.IsDigit(l.peek()) {
		l.advance()
	}

	// Treat a dot as part of the number only when it starts a fractional part.
	// This keeps the clause terminator in `value(2).` separate from the number.
	if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
		l.advance()

		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}

	return string(l.input[start:l.pos])
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}

	return l.input[l.pos]
}

func (l *lexer) peekNext() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}

	return l.input[l.pos+1]
}

func (l *lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}

	ch := l.input[l.pos]
	l.pos++

	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	return ch
}

func isIdentifierStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentifierPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}
