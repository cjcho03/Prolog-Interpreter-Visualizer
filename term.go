package prolog

import "strings"

type Term interface {
	String() string
}

type Atom string

func (a Atom) String() string {
	return string(a)
}

type Var string

func (v Var) String() string {
	text := string(v)

	if isAnonymousVar(v) {
		return "_"
	}

	return text
}

type Predicate struct {
	Name string
	Args []Term
}

func (p Predicate) String() string {
	args := make([]string, len(p.Args))

	for i, arg := range p.Args {
		args[i] = arg.String()
	}

	return p.Name + "(" + strings.Join(args, ", ") + ")"
}

type Substitution map[Var]Term

func copySubstitution(s Substitution) Substitution {
	result := make(Substitution)

	for key, value := range s {
		result[key] = value
	}

	return result
}

func isAnonymousVar(variable Var) bool {
	return strings.Contains(string(variable), "$anon_")
}
