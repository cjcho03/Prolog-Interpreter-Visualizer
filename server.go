package prolog

import (
	"encoding/json"
	"net/http"
)

type SolveResponse struct {
	Events  []TraceEvent        `json:"events"`
	Answers []map[string]string `json:"answers"`
}

func DemoHandler(engine Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := []Predicate{
			{
				Name: "parent",
				Args: []Term{Atom("alice"), Var("X")},
			},
			{
				Name: "parent",
				Args: []Term{Var("X"), Var("Y")},
			},
		}

		var events []TraceEvent

		answers := engine.SolveWithTrace(func(event TraceEvent) {
			events = append(events, event)
		}, query...)

		answerMaps := make([]map[string]string, len(answers))

		for i, answer := range answers {
			answerMaps[i] = snapshotBindings(answer)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(SolveResponse{
			Events:  events,
			Answers: answerMaps,
		})
	}
}
