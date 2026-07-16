package prolog

func dereference(term Term, sub Substitution) Term {
	for {
		variable, ok := term.(Var)
		if !ok {
			return term
		}

		value, found := sub[variable]
		if !found {
			return term
		}

		term = value
	}
}

func unify(left, right Term, sub Substitution) bool {
	left = dereference(left, sub)
	right = dereference(right, sub)

	if left == right {
		return true
	}

	if leftVar, ok := left.(Var); ok {
		sub[leftVar] = right
		return true
	}

	if rightVar, ok := right.(Var); ok {
		sub[rightVar] = left
		return true
	}

	leftAtom, leftIsAtom := left.(Atom)
	rightAtom, rightIsAtom := right.(Atom)

	if leftIsAtom && rightIsAtom {
		return leftAtom == rightAtom
	}

	leftNumber, leftIsNumber := left.(Number)
	rightNumber, rightIsNumber := right.(Number)

	return leftIsNumber && rightIsNumber && leftNumber == rightNumber
}

func unifyPredicate(goal Predicate, fact Predicate, sub Substitution) bool {
	if goal.Name != fact.Name || len(goal.Args) != len(fact.Args) {
		return false
	}

	for i := range goal.Args {
		if !unify(goal.Args[i], fact.Args[i], sub) {
			return false
		}
	}

	return true
}

func resolvePredicate(predicate Predicate, sub Substitution) Predicate {
	args := make([]Term, len(predicate.Args))

	for i, arg := range predicate.Args {
		args[i] = dereference(arg, sub)
	}

	return Predicate{
		Name: predicate.Name,
		Args: args,
	}
}
