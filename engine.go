package prolog

import (
	"math"
	"strconv"
	"strings"
)

// builtinPredicate executes a deterministic built-in goal.
// It may add bindings to sub and returns whether the goal succeeds.
type builtinPredicate func(args []Term, sub Substitution) bool

// predicateIndicator identifies a predicate by the standard name/arity pair.
type predicateIndicator struct {
	Name  string
	Arity int
}

// builtinPredicates is checked before user-defined clauses.
// Add future built-ins here by predicate name and arity.
var builtinPredicates = map[predicateIndicator]builtinPredicate{
	{Name: "true", Arity: 0}: builtinTrue,
	{Name: "fail", Arity: 0}: builtinFail,
	{Name: "=", Arity: 2}:    builtinUnify,
	{Name: "\\=", Arity: 2}:  builtinNotUnifiable,
	{Name: "is", Arity: 2}:   builtinIs,
	{Name: "<", Arity: 2}:    builtinLess,
	{Name: ">", Arity: 2}:    builtinGreater,
	{Name: "=<", Arity: 2}:   builtinLessEqual,
	{Name: ">=", Arity: 2}:   builtinGreaterEqual,
	{Name: "=:=", Arity: 2}:  builtinArithmeticEqual,
	{Name: "=\\=", Arity: 2}: builtinArithmeticNotEqual,
}

func builtinTrue(_ []Term, _ Substitution) bool {
	return true
}

func builtinFail(_ []Term, _ Substitution) bool {
	return false
}

func builtinUnify(args []Term, sub Substitution) bool {
	return unify(args[0], args[1], sub)
}

// builtinNotUnifiable succeeds only when its arguments cannot be unified.
// It probes unification against a copy so a failed \=/2 check never leaks
// temporary bindings into the current solution.
func builtinNotUnifiable(args []Term, sub Substitution) bool {
	probe := copySubstitution(sub)
	return !unify(args[0], args[1], probe)
}

type arithmeticValue struct {
	value   float64
	isFloat bool
}

func builtinIs(args []Term, sub Substitution) bool {
	value, ok := evaluateArithmetic(args[1], sub)
	if !ok {
		return false
	}

	return unify(args[0], Number(formatArithmeticValue(value)), sub)
}

func builtinLess(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left < right
	})
}

func builtinGreater(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left > right
	})
}

func builtinLessEqual(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left <= right
	})
}

func builtinGreaterEqual(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left >= right
	})
}

func builtinArithmeticEqual(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left == right
	})
}

func builtinArithmeticNotEqual(args []Term, sub Substitution) bool {
	return compareArithmetic(args, sub, func(left, right float64) bool {
		return left != right
	})
}

func compareArithmetic(
	args []Term,
	sub Substitution,
	compare func(left, right float64) bool,
) bool {
	left, ok := evaluateArithmetic(args[0], sub)
	if !ok {
		return false
	}

	right, ok := evaluateArithmetic(args[1], sub)
	if !ok {
		return false
	}

	return compare(left.value, right.value)
}

func evaluateArithmetic(term Term, sub Substitution) (arithmeticValue, bool) {
	term = resolveTerm(term, sub)

	switch value := term.(type) {
	case Number:
		parsed, err := strconv.ParseFloat(string(value), 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return arithmeticValue{}, false
		}

		return arithmeticValue{
			value:   parsed,
			isFloat: strings.Contains(string(value), "."),
		}, true

	case Compound:
		return evaluateArithmeticCompound(value, sub)

	default:
		// Unbound variables and non-numeric atoms are not evaluable.
		return arithmeticValue{}, false
	}
}

func evaluateArithmeticCompound(
	compound Compound,
	sub Substitution,
) (arithmeticValue, bool) {
	if compound.Name == "abs" && len(compound.Args) == 1 {
		operand, ok := evaluateArithmetic(compound.Args[0], sub)
		if !ok {
			return arithmeticValue{}, false
		}

		operand.value = math.Abs(operand.value)
		return operand, true
	}

	if (compound.Name == "+" || compound.Name == "-") && len(compound.Args) == 1 {
		operand, ok := evaluateArithmetic(compound.Args[0], sub)
		if !ok {
			return arithmeticValue{}, false
		}

		if compound.Name == "-" {
			operand.value = -operand.value
		}

		return operand, true
	}

	if len(compound.Args) != 2 {
		return arithmeticValue{}, false
	}

	left, ok := evaluateArithmetic(compound.Args[0], sub)
	if !ok {
		return arithmeticValue{}, false
	}

	right, ok := evaluateArithmetic(compound.Args[1], sub)
	if !ok {
		return arithmeticValue{}, false
	}

	if compound.Name == "min" {
		if left.value <= right.value {
			return left, true
		}

		return right, true
	}

	if compound.Name == "max" {
		if left.value >= right.value {
			return left, true
		}

		return right, true
	}

	result := arithmeticValue{
		isFloat: left.isFloat || right.isFloat,
	}

	switch compound.Name {
	case "+":
		result.value = left.value + right.value
	case "-":
		result.value = left.value - right.value
	case "*":
		result.value = left.value * right.value
	case "/":
		if right.value == 0 {
			return arithmeticValue{}, false
		}

		result.value = left.value / right.value
		result.isFloat = true
	case "mod":
		if left.isFloat || right.isFloat || right.value == 0 {
			return arithmeticValue{}, false
		}

		result.value = math.Mod(left.value, right.value)
	case "**":
		result.value = math.Pow(left.value, right.value)
		if right.value < 0 || math.Trunc(result.value) != result.value {
			result.isFloat = true
		}
	default:
		return arithmeticValue{}, false
	}

	if math.IsNaN(result.value) || math.IsInf(result.value, 0) {
		return arithmeticValue{}, false
	}

	return result, true
}

func formatArithmeticValue(value arithmeticValue) string {
	text := strconv.FormatFloat(value.value, 'f', -1, 64)

	if value.isFloat && !strings.Contains(text, ".") {
		return text + ".0"
	}

	return text
}

func lookupBuiltin(goal Predicate) (builtinPredicate, bool) {
	builtin, found := builtinPredicates[predicateIndicator{
		Name:  goal.Name,
		Arity: len(goal.Args),
	}]

	return builtin, found
}

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

		if builtin, found := lookupBuiltin(currentGoal); found {
			emit(sink, TraceEvent{
				Type:        EventBuiltin,
				Depth:       depth,
				Goal:        currentGoal.String(),
				Bindings:    snapshotQueryBindings(sub, userQueryVars),
				Description: "Executing built-in predicate.",
			})

			nextSub := copySubstitution(sub)

			if !builtin(currentGoal.Args, nextSub) {
				emit(sink, TraceEvent{
					Type:        EventFailed,
					Depth:       depth,
					Goal:        currentGoal.String(),
					Bindings:    snapshotQueryBindings(nextSub, userQueryVars),
					Description: "Built-in predicate failed.",
				})

				emit(sink, TraceEvent{
					Type:        EventBacktrack,
					Depth:       depth,
					Goal:        currentGoal.String(),
					Bindings:    snapshotQueryBindings(sub, userQueryVars),
					Description: "Built-in predicate produced no solution. Returning to the previous decision.",
				})

				return false
			}

			return search(remainingAfterGoal, nextSub, depth+1)
		}

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
