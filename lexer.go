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
	tokenComma
	tokenDot
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
	case tokenDot:
		return "."
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

	return string(l.input[start:l.pos])
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}

	return l.input[l.pos]
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
