package prolog

import (
	"fmt"
	"strings"
)

type Clause struct {
	Head Predicate
	Body []Predicate
}

func Fact(head Predicate) Clause {
	return Clause{
		Head: head,
	}
}

func Rule(head Predicate, body ...Predicate) Clause {
	copiedBody := make([]Predicate, len(body))
	copy(copiedBody, body)

	return Clause{
		Head: head,
		Body: copiedBody,
	}
}

func (c Clause) IsFact() bool {
	return len(c.Body) == 0
}

func (c Clause) String() string {
	if c.IsFact() {
		return c.Head.String()
	}

	body := make([]string, len(c.Body))

	for i, goal := range c.Body {
		body[i] = goal.String()
	}

	return c.Head.String() + " :- " + strings.Join(body, ", ")
}

// standardizeApart gives every use of a clause its own variables
// For example, using ancestory(X, Y) twice should create two independent
// sets of internal variables rather than sharing X and Y across both calls
func standardizeApart(clause Clause, id int) Clause {
	renamed := make(map[Var]Var)

	renameTerm := func(term Term) Term {
		variable, ok := term.(Var)
		if !ok {
			return term
		}

		fresh, found := renamed[variable]
		if !found {
			fresh = Var(fmt.Sprintf("$%d_%s", id, variable))
			renamed[variable] = fresh
		}

		return fresh
	}

	renamePredicate := func(predicate Predicate) Predicate {
		args := make([]Term, len(predicate.Args))

		for i, arg := range predicate.Args {
			args[i] = renameTerm(arg)
		}

		return Predicate{
			Name: predicate.Name,
			Args: args,
		}
	}

	body := make([]Predicate, len(clause.Body))

	for i, goal := range clause.Body {
		body[i] = renamePredicate(goal)
	}

	return Clause{
		Head: renamePredicate(clause.Head),
		Body: body,
	}
}
