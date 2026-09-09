package prolog

import "fmt"

type parser struct {
	tokens          []token
	pos             int
	nextAnonymousID int
}

// ParseProgram parses Prolog facts and rules.
// Examples:
// parent(alice, bob).
// grandparent(X, Z) :- parent(X, Y), parent(Y, Z).
func ParseProgram(src string) ([]Clause, error) {
	tokens, err := lex(src)
	if err != nil {
		return nil, err
	}

	p := parser{tokens: tokens}

	clauses, err := p.parseProgram()
	if err != nil {
		return nil, err
	}

	if err := p.expect(tokenEOF); err != nil {
		return nil, err
	}

	return clauses, nil
}

// ParseQuery parses a Prolog query.
// Examples:
// ?- parent(alice, Who).
// ?- parent(alice, X), parent(X, Y).
func ParseQuery(src string) ([]Predicate, error) {
	tokens, err := lex(src)
	if err != nil {
		return nil, err
	}

	p := parser{tokens: tokens}

	if err := p.expect(tokenQuery); err != nil {
		return nil, err
	}

	goals, err := p.parseGoalList()
	if err != nil {
		return nil, err
	}

	if err := p.expect(tokenDot); err != nil {
		return nil, err
	}

	if err := p.expect(tokenEOF); err != nil {
		return nil, err
	}

	return goals, nil
}

func (p *parser) parseProgram() ([]Clause, error) {
	var clauses []Clause

	for p.peek().typ != tokenEOF {
		if p.peek().typ == tokenQuery {
			return nil, p.errorf(
				p.peek(),
				"queries are not allowed in ParseProgram: use ParseQuery instead",
			)
		}

		clause, err := p.parseClause()
		if err != nil {
			return nil, err
		}

		clauses = append(clauses, clause)
	}

	return clauses, nil
}

func (p *parser) parseClause() (Clause, error) {
	head, err := p.parsePredicate()
	if err != nil {
		return Clause{}, err
	}

	switch p.peek().typ {
	case tokenDot:
		p.advance()
		return Fact(head), nil
	case tokenColonDash:
		p.advance()

		body, err := p.parseGoalList()
		if err != nil {
			return Clause{}, err
		}

		if err := p.expect(tokenDot); err != nil {
			return Clause{}, err
		}

		return Rule(head, body...), nil
	default:
		return Clause{}, p.errorf(
			p.peek(),
			"expected '.' for a fact or ':-' for a rule",
		)
	}
}

func (p *parser) parseGoalList() ([]Predicate, error) {
	first, err := p.parseGoal()
	if err != nil {
		return nil, err
	}

	goals := []Predicate{first}

	for p.match(tokenComma) {
		next, err := p.parseGoal()
		if err != nil {
			return nil, err
		}

		goals = append(goals, next)
	}

	return goals, nil
}

func (p *parser) parseGoal() (Predicate, error) {
	left, err := p.parseTerm()
	if err != nil {
		return Predicate{}, err
	}

	switch p.peek().typ {
	case tokenEquals,
		tokenLess,
		tokenGreater,
		tokenLessEqual,
		tokenGreaterEqual,
		tokenArithmeticEqual,
		tokenArithmeticNotEqual:
		op := p.advance().literal

		right, err := p.parseTerm()
		if err != nil {
			return Predicate{}, err
		}

		return Predicate{
			Name: op,
			Args: []Term{left, right},
		}, nil
	}

	if p.peek().typ == tokenIdent && p.peek().literal == "is" {
		p.advance()

		right, err := p.parseTerm()
		if err != nil {
			return Predicate{}, err
		}

		return Predicate{
			Name: "is",
			Args: []Term{left, right},
		}, nil
	}

	switch value := left.(type) {
	case Atom:
		return Predicate{Name: string(value)}, nil
	case Compound:
		return Predicate{
			Name: value.Name,
			Args: value.Args,
		}, nil
	default:
		return Predicate{}, p.errorf(
			p.peek(),
			"expected predicate goal, comparison operator, '=', or 'is' after term",
		)
	}
}

func (p *parser) parsePredicate() (Predicate, error) {
	name, err := p.consume(tokenIdent, "expected predicate name")
	if err != nil {
		return Predicate{}, err
	}

	predicate := Predicate{
		Name: name.literal,
	}

	if !p.match(tokenLParen) {
		return predicate, nil
	}

	// Allows zero-arity syntax like foo(), even though plain is preferred.
	if p.match(tokenRParen) {
		return predicate, nil
	}

	args, err := p.parseTermList()
	if err != nil {
		return Predicate{}, err
	}
	if err := p.expect(tokenRParen); err != nil {
		return Predicate{}, err
	}

	predicate.Args = args
	return predicate, nil
}

func (p *parser) parseTermList() ([]Term, error) {
	first, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	terms := []Term{first}

	for p.match(tokenComma) {
		next, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		terms = append(terms, next)
	}

	return terms, nil
}

func (p *parser) parseTerm() (Term, error) {
	return p.parseAdditiveTerm()
}

// Arithmetic operators are represented as ordinary compound terms. For example,
// 2 + 3 * 4 parses as +(2, *(3, 4)). Evaluation happens only through is/2.
func (p *parser) parseAdditiveTerm() (Term, error) {
	left, err := p.parseMultiplicativeTerm()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokenPlus || p.peek().typ == tokenMinus {
		op := p.advance().literal

		right, err := p.parseMultiplicativeTerm()
		if err != nil {
			return nil, err
		}

		left = Compound{
			Name: op,
			Args: []Term{left, right},
		}
	}

	return left, nil
}

func (p *parser) parseMultiplicativeTerm() (Term, error) {
	left, err := p.parseUnaryTerm()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokenStar ||
		p.peek().typ == tokenSlash ||
		(p.peek().typ == tokenIdent && p.peek().literal == "mod") {
		op := p.advance().literal

		right, err := p.parseUnaryTerm()
		if err != nil {
			return nil, err
		}

		left = Compound{
			Name: op,
			Args: []Term{left, right},
		}
	}

	return left, nil
}

func (p *parser) parseUnaryTerm() (Term, error) {
	if p.peek().typ == tokenPlus || p.peek().typ == tokenMinus {
		op := p.advance().literal

		operand, err := p.parseUnaryTerm()
		if err != nil {
			return nil, err
		}

		return Compound{
			Name: op,
			Args: []Term{operand},
		}, nil
	}

	return p.parsePowerTerm()
}

// Exponentiation binds more tightly than unary +/- and is right-associative.
// For example, 2 ** 3 ** 2 parses as **(2, **(3, 2)).
func (p *parser) parsePowerTerm() (Term, error) {
	left, err := p.parsePrimaryTerm()
	if err != nil {
		return nil, err
	}

	if !p.match(tokenPower) {
		return left, nil
	}

	right, err := p.parseUnaryTerm()
	if err != nil {
		return nil, err
	}

	return Compound{
		Name: "**",
		Args: []Term{left, right},
	}, nil
}

func (p *parser) parsePrimaryTerm() (Term, error) {
	tok := p.peek()

	switch tok.typ {
	case tokenIdent:
		p.advance()

		if !p.match(tokenLParen) {
			return Atom(tok.literal), nil
		}

		if p.match(tokenRParen) {
			return Compound{Name: tok.literal}, nil
		}

		args, err := p.parseTermList()
		if err != nil {
			return nil, err
		}

		if err := p.expect(tokenRParen); err != nil {
			return nil, err
		}

		return Compound{
			Name: tok.literal,
			Args: args,
		}, nil

	case tokenNumber:
		p.advance()
		return Number(tok.literal), nil

	case tokenVar:
		p.advance()

		if tok.literal == "_" {
			return p.freshAnonymousVar(), nil
		}

		return Var(tok.literal), nil

	case tokenLBracket:
		return p.parseListTerm()

	case tokenLParen:
		p.advance()

		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		if err := p.expect(tokenRParen); err != nil {
			return nil, err
		}

		return term, nil

	default:
		return nil, p.errorf(tok, "expected atom, number, variable, compound term, list, or parenthesized term")
	}
}

func (p *parser) parseListTerm() (Term, error) {
	if err := p.expect(tokenLBracket); err != nil {
		return nil, err
	}

	if p.match(tokenRBracket) {
		return Atom("[]"), nil
	}

	first, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	elements := []Term{first}
	tail := Term(Atom("[]"))

	for p.match(tokenComma) {
		next, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		elements = append(elements, next)
	}

	if p.match(tokenBar) {
		tail, err = p.parseTerm()
		if err != nil {
			return nil, err
		}
	}

	if err := p.expect(tokenRBracket); err != nil {
		return nil, err
	}

	return buildList(elements, tail), nil
}

func buildList(elements []Term, tail Term) Term {
	result := tail

	for i := len(elements) - 1; i >= 0; i-- {
		result = Compound{
			Name: ".",
			Args: []Term{
				elements[i],
				result,
			},
		}
	}

	return result
}

func (p *parser) freshAnonymousVar() Var {
	p.nextAnonymousID++
	return Var(fmt.Sprintf("$anon_%d", p.nextAnonymousID))
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokenEOF}
	}

	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	tok := p.peek()

	if p.pos < len(p.tokens) {
		p.pos++
	}

	return tok
}

func (p *parser) match(want tokenType) bool {
	if p.peek().typ != want {
		return false
	}

	p.advance()
	return true
}

func (p *parser) expect(want tokenType) error {
	_, err := p.consume(want, fmt.Sprintf("expected %s", tokenName(want)))
	return err
}

func (p *parser) consume(want tokenType, message string) (token, error) {
	tok := p.peek()

	if tok.typ != want {
		return token{}, p.errorf(tok, message)
	}

	p.advance()
	return tok, nil
}

func (p *parser) errorf(tok token, message string) error {
	return fmt.Errorf(
		"%d:%d: %s, got %s",
		tok.line,
		tok.column,
		message,
		tok.String(),
	)
}

func tokenName(typ tokenType) string {
	switch typ {
	case tokenEOF:
		return "EOF"
	case tokenIdent:
		return "identifier"
	case tokenNumber:
		return "number"
	case tokenVar:
		return "variable"
	case tokenLParen:
		return "("
	case tokenRParen:
		return ")"
	case tokenLBracket:
		return "["
	case tokenRBracket:
		return "]"
	case tokenComma:
		return ","
	case tokenDot:
		return "."
	case tokenBar:
		return "|"
	case tokenEquals:
		return "="
	case tokenPlus:
		return "+"
	case tokenMinus:
		return "-"
	case tokenStar:
		return "*"
	case tokenSlash:
		return "/"
	case tokenColonDash:
		return ":-"
	case tokenQuery:
		return "?-"
	default:
		return "token"
	}
}
