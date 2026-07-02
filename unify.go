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

	// X = X should succeed without storing X -> X
	if left == right {
		return true
	}

	if leftAtom, ok := left.(Atom); ok {
		rightAtom, ok := right.(Atom)
		return ok && leftAtom == rightAtom
	}

	if leftVar, ok := left.(Var); ok {
		sub[leftVar] = right
		return true
	}

	if rightVar, ok := right.(Var); ok {
		sub[rightVar] = left
		return true
	}

	return false
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
