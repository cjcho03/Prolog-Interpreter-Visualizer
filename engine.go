package prolog

type Engine struct {
	Facts []Predicate
}

func (e Engine) Solve(goals ...Predicate) []Substitution {
	return e.SolveWithTrace(nil, goals...)
}
func (e Engine) SolveWithTrace(sink TraceSink, goals ...Predicate) []Substitution {
	var results []Substitution

	// Returns true when this branch produces at least one solution
	var search func(goalIndex int, sub Substitution) bool

	search = func(goalIndex int, sub Substitution) bool {
		if goalIndex == len(goals) {
			emit(sink, TraceEvent{
				Type:        EventSolution,
				Depth:       goalIndex,
				Bindings:    snapshotBindings(sub),
				Description: "All goals matched. Solution found.",
			})

			results = append(results, copySubstitution(sub))
			return true
		}

		currentGoal := goals[goalIndex]

		emit(sink, TraceEvent{
			Type:        EventGoal,
			Depth:       goalIndex,
			Goal:        currentGoal.String(),
			Bindings:    snapshotBindings(sub),
			Description: "Trying to satisfy the next goal.",
		})

		foundMatch := false
		foundSolution := false

		for _, fact := range e.Facts {
			nextSub := copySubstitution(sub)

			emit(sink, TraceEvent{
				Type:        EventTryFact,
				Depth:       goalIndex,
				Goal:        currentGoal.String(),
				Fact:        fact.String(),
				Bindings:    snapshotBindings(nextSub),
				Description: "Trying this fact against the current goal.",
			})

			if !unifyPredicate(currentGoal, fact, nextSub) {
				emit(sink, TraceEvent{
					Type:        EventFailed,
					Depth:       goalIndex,
					Goal:        currentGoal.String(),
					Fact:        fact.String(),
					Bindings:    snapshotBindings(nextSub),
					Description: "This fact does not unify with the goal.",
				})
				continue
			}

			foundMatch = true
			emit(sink, TraceEvent{
				Type:        EventUnified,
				Depth:       goalIndex,
				Goal:        currentGoal.String(),
				Fact:        fact.String(),
				Bindings:    snapshotBindings(nextSub),
				Description: "Unification succeeded.",
			})
			branchSucceeded := search(goalIndex+1, nextSub)

			if !branchSucceeded {
				emit(sink, TraceEvent{
					Type:        EventBacktrack,
					Depth:       goalIndex,
					Goal:        currentGoal.String(),
					Fact:        fact.String(),
					Bindings:    snapshotBindings(nextSub),
					Description: "This branch produced no solution. Backtracking to try another fact.",
				})
			} else {
				foundSolution = true
			}
		}

		if !foundMatch {
			emit(sink, TraceEvent{
				Type:        EventBacktrack,
				Depth:       goalIndex,
				Goal:        currentGoal.String(),
				Bindings:    snapshotBindings(sub),
				Description: "No remaining facts match this goal. Returning to the previous decision.",
			})
		}
		return foundSolution
	}

	search(0, Substitution{})
	return results
}
