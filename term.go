package prolog

import "strings"

type Term interface {
	String() string
}

type Atom string

func (a Atom) String() string {
	return string(a)
}

type Number string

func (n Number) String() string {
	return string(n)
}

type Var string

func (v Var) String() string {
	text := string(v)

	if isAnonymousVar(v) {
		return "_"
	}

	return text
}

type Compound struct {
	Name string
	Args []Term
}

func (c Compound) String() string {
	if c.Name == "." && len(c.Args) == 2 {
		elements, tail, ok := listParts(c)
		if ok {
			return formatList(elements, tail)
		}
	}

	if isArithmeticCompound(c) {
		return formatArithmeticTerm(c, 0, false)
	}

	return formatTermWithArgs(c.Name, c.Args)
}

type Predicate struct {
	Name string
	Args []Term
}

func (p Predicate) String() string {
	if len(p.Args) == 2 && isInfixPredicate(p.Name) {
		return p.Args[0].String() + " " + p.Name + " " + p.Args[1].String()
	}

	return formatTermWithArgs(p.Name, p.Args)
}

func isInfixPredicate(name string) bool {
	switch name {
	case "=", "\\=", "is", "<", ">", "=<", ">=", "=:=", "=\\=":
		return true
	default:
		return false
	}
}

type Substitution map[Var]Term

func copySubstitution(s Substitution) Substitution {
	result := make(Substitution)

	for key, value := range s {
		result[key] = value
	}

	return result
}

func isArithmeticCompound(compound Compound) bool {
	if len(compound.Args) == 1 {
		return compound.Name == "+" || compound.Name == "-"
	}

	if len(compound.Args) == 2 {
		switch compound.Name {
		case "+", "-", "*", "/", "mod", "**":
			return true
		}
	}

	return false
}

func arithmeticPrecedence(compound Compound) int {
	if len(compound.Args) == 1 {
		return 3
	}

	switch compound.Name {
	case "+", "-":
		return 1
	case "*", "/", "mod":
		return 2
	case "**":
		return 4
	default:
		return 5
	}
}

func formatArithmeticTerm(term Term, parentPrecedence int, rightChild bool) string {
	return formatArithmeticTermWithParent(term, parentPrecedence, rightChild, "")
}

func formatArithmeticTermWithParent(
	term Term,
	parentPrecedence int,
	rightChild bool,
	parentOperator string,
) string {
	compound, ok := term.(Compound)
	if !ok || !isArithmeticCompound(compound) {
		return term.String()
	}

	precedence := arithmeticPrecedence(compound)
	var text string

	if len(compound.Args) == 1 {
		operand := formatArithmeticTermWithParent(
			compound.Args[0],
			precedence,
			true,
			compound.Name,
		)
		text = compound.Name + operand
	} else {
		left := formatArithmeticTermWithParent(
			compound.Args[0],
			precedence,
			false,
			compound.Name,
		)
		right := formatArithmeticTermWithParent(
			compound.Args[1],
			precedence,
			true,
			compound.Name,
		)
		text = left + " " + compound.Name + " " + right
	}

	needsParentheses := precedence < parentPrecedence

	if precedence == parentPrecedence {
		if parentOperator == "**" {
			// ** is right-associative, so a same-precedence left child must have
			// come from explicit grouping.
			needsParentheses = !rightChild
		} else {
			// The other binary operators are left-associative. A same-precedence
			// right child therefore needs parentheses to preserve the expression tree.
			needsParentheses = rightChild
		}
	}

	if needsParentheses {
		return "(" + text + ")"
	}

	return text
}

func formatTermWithArgs(name string, args []Term) string {
	formattedArgs := make([]string, len(args))

	for i, arg := range args {
		formattedArgs[i] = arg.String()
	}

	return name + "(" + strings.Join(formattedArgs, ", ") + ")"
}

func listParts(term Term) ([]Term, Term, bool) {
	var elements []Term
	current := term

	for {
		if isEmptyList(current) {
			return elements, Atom("[]"), true
		}

		compound, ok := current.(Compound)
		if !ok || compound.Name != "." || len(compound.Args) != 2 {
			return elements, current, true
		}

		elements = append(elements, compound.Args[0])
		current = compound.Args[1]
	}
}

func formatList(elements []Term, tail Term) string {
	formattedElements := make([]string, len(elements))

	for i, element := range elements {
		formattedElements[i] = element.String()
	}

	if !isEmptyList(tail) {
		return "[" + strings.Join(formattedElements, ", ") + " | " + tail.String() + "]"
	}

	return "[" + strings.Join(formattedElements, ", ") + "]"
}

func isEmptyList(term Term) bool {
	atom, ok := term.(Atom)
	return ok && atom == Atom("[]")
}

func isAnonymousVar(variable Var) bool {
	return strings.Contains(string(variable), "$anon_")
}
