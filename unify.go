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

	if leftVar, ok := left.(Var); ok {
		if rightVar, ok := right.(Var); ok && leftVar == rightVar {
			return true
		}

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

	if leftIsNumber && rightIsNumber {
		return leftNumber == rightNumber
	}

	leftCompound, leftIsCompound := left.(Compound)
	rightCompound, rightIsCompound := right.(Compound)

	if leftIsCompound && rightIsCompound {
		return unifyCompound(leftCompound, rightCompound, sub)
	}

	return false
}

func unifyCompound(left, right Compound, sub Substitution) bool {
	if left.Name != right.Name || len(left.Args) != len(right.Args) {
		return false
	}

	for i := range left.Args {
		if !unify(left.Args[i], right.Args[i], sub) {
			return false
		}
	}

	return true
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

func resolveTerm(term Term, sub Substitution) Term {
	term = dereference(term, sub)

	compound, ok := term.(Compound)
	if !ok {
		return term
	}

	args := make([]Term, len(compound.Args))

	for i, arg := range compound.Args {
		args[i] = resolveTerm(arg, sub)
	}

	return Compound{
		Name: compound.Name,
		Args: args,
	}
}

func resolvePredicate(predicate Predicate, sub Substitution) Predicate {
	args := make([]Term, len(predicate.Args))

	for i, arg := range predicate.Args {
		args[i] = resolveTerm(arg, sub)
	}

	return Predicate{
		Name: predicate.Name,
		Args: args,
	}
}
