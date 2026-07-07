package prolog

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

func snapshotBindings(sub Substitution) map[string]string {
	result := make(map[string]string)

	for variable, value := range sub {
		result[string(variable)] = dereference(value, sub).String()
	}

	return result
}
