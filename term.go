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

func isAnonymousVar(variable Var) bool {
	return strings.Contains(string(variable), "$anon_")
}
