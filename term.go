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
	return formatTermWithArgs(c.Name, c.Args)
}

type Predicate struct {
	Name string
	Args []Term
}

func (p Predicate) String() string {
	return formatTermWithArgs(p.Name, p.Args)
}

type Substitution map[Var]Term

func copySubstitution(s Substitution) Substitution {
	result := make(Substitution)

	for key, value := range s {
		result[key] = value
	}

	return result
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
