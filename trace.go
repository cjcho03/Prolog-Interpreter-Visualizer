package prolog

import "strings"

type TraceEventType string

const (
	EventGoal         TraceEventType = "goal"
	EventBuiltin      TraceEventType = "builtin"
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

func queryVariables(goals []Predicate) []Var {
	seen := make(map[Var]bool)
	var variables []Var

	var collect func(term Term)

	collect = func(term Term) {
		switch value := term.(type) {
		case Var:
			if isInternalVar(value) || seen[value] {
				return
			}

			seen[value] = true
			variables = append(variables, value)

		case Compound:
			for _, arg := range value.Args {
				collect(arg)
			}
		}
	}

	for _, goal := range goals {
		for _, arg := range goal.Args {
			collect(arg)
		}
	}

	return variables
}

func snapshotBindings(sub Substitution) map[string]string {
	result := make(map[string]string)

	for variable, value := range sub {
		result[string(variable)] = resolveTerm(value, sub).String()
	}

	return result
}

func snapshotQueryBindings(sub Substitution, queryVars []Var) map[string]string {
	result := make(map[string]string)

	for _, variable := range queryVars {
		value, found := sub[variable]
		if !found {
			continue
		}

		resolved := resolveTerm(value, sub)

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
		resolved := resolveTerm(variable, sub)

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
