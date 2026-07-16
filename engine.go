package prolog

type Engine struct {
	Clauses []Clause
}

func (e Engine) Solve(goals ...Predicate) []Substitution {
	return e.SolveWithTrace(nil, goals...)
}

func (e Engine) SolveWithTrace(
	sink TraceSink,
	goals ...Predicate,
) []Substitution {
	var results []Substitution

	userQueryVars := queryVariables(goals)

	// Gives each clause use a distinct set of internal variables.
	nextClauseID := 1

	// Returns true when this path produces at least one solution.
	var search func(
		remainingGoals []Predicate,
		sub Substitution,
		depth int,
	) bool

	search = func(
		remainingGoals []Predicate,
		sub Substitution,
		depth int,
	) bool {
		if len(remainingGoals) == 0 {
			emit(sink, TraceEvent{
				Type:        EventSolution,
				Depth:       depth,
				Bindings:    snapshotQueryBindings(sub, userQueryVars),
				Description: "All goals matched. Solution found.",
			})

			results = append(results, copySubstitution(sub))
			return true
		}

		currentGoal := resolvePredicate(remainingGoals[0], sub)
		remainingAfterGoal := remainingGoals[1:]

		emit(sink, TraceEvent{
			Type:        EventGoal,
			Depth:       depth,
			Goal:        currentGoal.String(),
			Bindings:    snapshotQueryBindings(sub, userQueryVars),
			Description: "Trying to satisfy the next goal.",
		})

		foundMatch := false
		foundSolution := false

		for _, clause := range e.Clauses {
			freshClause := standardizeApart(clause, nextClauseID)
			nextClauseID++

			nextSub := copySubstitution(sub)

			emit(sink, TraceEvent{
				Type:        EventTryClause,
				Depth:       depth,
				Goal:        currentGoal.String(),
				Clause:      clause.String(),
				Bindings:    snapshotQueryBindings(nextSub, userQueryVars),
				Description: "Trying this clause against the current goal.",
			})

			if !unifyPredicate(currentGoal, freshClause.Head, nextSub) {
				emit(sink, TraceEvent{
					Type:        EventFailed,
					Depth:       depth,
					Goal:        currentGoal.String(),
					Clause:      clause.String(),
					Bindings:    snapshotQueryBindings(nextSub, userQueryVars),
					Description: "This clause does not unify with the goal.",
				})

				continue
			}

			foundMatch = true

			emit(sink, TraceEvent{
				Type:        EventUnified,
				Depth:       depth,
				Goal:        currentGoal.String(),
				Clause:      clause.String(),
				Bindings:    snapshotQueryBindings(nextSub, userQueryVars),
				Description: "Unification succeeded.",
			})

			if !freshClause.IsFact() {
				emit(sink, TraceEvent{
					Type:          EventRuleExpanded,
					Depth:         depth,
					Goal:          currentGoal.String(),
					Clause:        clause.String(),
					ExpandedGoals: resolvedGoalStrings(freshClause.Body, nextSub),
					Bindings:      snapshotQueryBindings(nextSub, userQueryVars),
					Description:   "Rule matched. Expanding its body into the next goals.",
				})
			}

			nextGoals := make(
				[]Predicate,
				0,
				len(freshClause.Body)+len(remainingAfterGoal),
			)

			nextGoals = append(nextGoals, freshClause.Body...)
			nextGoals = append(nextGoals, remainingAfterGoal...)

			branchSucceeded := search(nextGoals, nextSub, depth+1)

			if !branchSucceeded {
				emit(sink, TraceEvent{
					Type:        EventBacktrack,
					Depth:       depth,
					Goal:        currentGoal.String(),
					Clause:      clause.String(),
					Bindings:    snapshotQueryBindings(nextSub, userQueryVars),
					Description: "This branch produced no solution. Backtracking to try another clause.",
				})
			} else {
				foundSolution = true
			}
		}

		if !foundMatch {
			emit(sink, TraceEvent{
				Type:        EventBacktrack,
				Depth:       depth,
				Goal:        currentGoal.String(),
				Bindings:    snapshotQueryBindings(sub, userQueryVars),
				Description: "No remaining clauses match this goal. Returning to the previous decision.",
			})
		}

		return foundSolution
	}

	search(goals, Substitution{}, 0)
	return results
}

func resolvedGoalStrings(goals []Predicate, sub Substitution) []string {
	result := make([]string, len(goals))

	for i, goal := range goals {
		result[i] = resolvePredicate(goal, sub).String()
	}

	return result
}
