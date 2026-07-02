package prolog

type Engine struct {
	Facts []Predicate
}

func (e Engine) Solve(goals ...Predicate) []Substitution {
	var results []Substitution

	var search func(goalIndex int, sub Substitution)

	search = func(goalIndex int, sub Substitution) {
		if goalIndex == len(goals) {
			results = append(results, sub)
			return
		}

		currentGoal := goals[goalIndex]

		for _, fact := range e.Facts {
			nextSub := copySubstitution(sub)

			if unifyPredicate(currentGoal, fact, nextSub) {
				search(goalIndex+1, nextSub)
			}
		}
	}

	search(0, Substitution{})
	return results
}
