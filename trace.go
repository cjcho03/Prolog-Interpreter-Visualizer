package prolog

import "strings"

type TraceEventType string

const (
	EventGoal         TraceEventType = "goal"
	EventTryClause    TraceEventType = "try_clause"
	EventUnified      TraceEventType = "unified"
	EventRuleExpanded TraceEventType = "rule_expanded"
	EventFailed       TraceEventType = "failed"
	EventBacktrack    TraceEventType = "backtrack"
	EventSolution     TraceEventType = "solution"
)

type TraceEvent struct {
	Type          TraceEventType    `json:"type"`
	Depth         int               `json:"depth"`
	Goal          string            `json:"goal,omitempty"`
	Clause        string            `json:"clause,omitempty"`
	ExpandedGoals []string          `json:"expandedGoals,omitempty"`
	Bindings      map[string]string `json:"bindings,omitempty"`
	Description   string            `json:"description"`
}

type TraceSink func(event TraceEvent)

func emit(sink TraceSink, event TraceEvent) {
	if sink != nil {
		sink(event)
	}
}

// queryVariables returns the user-facing variables that appeared in the
// original query. Internal standardized-apart variables are intentionally
// excluded from user-facing trace bindings and final answers.
func queryVariables(goals []Predicate) []Var {
	seen := make(map[Var]bool)
	var variables []Var

	for _, goal := range goals {
		for _, arg := range goal.Args {
			variable, ok := arg.(Var)
			if !ok {
				continue
			}

			if isInternalVar(variable) {
				continue
			}

			seen[variable] = true
			variables = append(variables, variable)
		}
	}

	return variables
}

func snapshotBindings(sub Substitution) map[string]string {
	result := make(map[string]string)

	for variable, value := range sub {
		result[string(variable)] = dereference(value, sub).String()
	}

	return result
}

// snapshotQueryBindings returns only bindings for variables the user actually
// asked about in the query.
// If a query variable is currently bound only to an unresolved internal rule
// variable, it is omitted for that trace step.
func snapshotQueryBindings(sub Substitution, queryVars []Var) map[string]string {
	result := make(map[string]string)

	for _, variable := range queryVars {
		value, found := sub[variable]
		if !found {
			continue
		}

		resolved := dereference(value, sub)

		if resolvedVar, ok := resolved.(Var); ok && isInternalVar(resolvedVar) {
			continue
		}

		result[string(variable)] = resolved.String()
	}

	return result
}

func answerBindings(sub Substitution, queryVars []Var) map[string]string {
	result := make(map[string]string)

	for _, variable := range queryVars {
		resolved := dereference(variable, sub)

		if resolvedVar, ok := resolved.(Var); ok && isInternalVar(resolvedVar) {
			continue
		}

		if resolved == variable {
			continue
		}

		result[string(variable)] = resolved.String()
	}

	return result
}

func isInternalVar(variable Var) bool {
	return strings.HasPrefix(string(variable), "$")
}
