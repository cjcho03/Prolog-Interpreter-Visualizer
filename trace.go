package prolog

type TraceEventType string

const (
	EventGoal      TraceEventType = "goal"
	EventTryFact   TraceEventType = "try_fact"
	EventUnified   TraceEventType = "unified"
	EventFailed    TraceEventType = "failed"
	EventBacktrack TraceEventType = "backtrack"
	EventSolution  TraceEventType = "solution"
)

type TraceEvent struct {
	Type        TraceEventType    `json:"type"`
	Depth       int               `json:"depth"`
	Goal        string            `json:"goal,omitempty"`
	Fact        string            `json:"fact,omitempty"`
	Bindings    map[string]string `json:"bindings,omitempty"`
	Description string            `json:"description"`
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
