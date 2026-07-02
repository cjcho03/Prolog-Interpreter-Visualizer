package prolog

import "fmt"

type Term interface {
	String() string
}

type Atom string

func (a Atom) String() string {
	return string(a)
}

type Var string

func (v Var) String() string {
	return string(v)
}

type Predicate struct {
	Name string
	Args []Term
}

func (p Predicate) String() string {
	return fmt.Sprintf("%s%v", p.Name, p.Args)
}

type Substitution map[Var]Term

func copySubstitution(s Substitution) Substitution {
	result := make(Substitution)

	for key, value := range s {
		result[key] = value
	}

	return result
}
