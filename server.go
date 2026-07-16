package prolog

import (
	"encoding/json"
	"net/http"
)

type SolveRequest struct {
	Program string `json:"program"`
	Query   string `json:"query"`
}

type SolveResponse struct {
	Events  []TraceEvent        `json:"events"`
	Answers []map[string]string `json:"answers"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func DemoHandler(engine Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		query := []Predicate{
			{
				Name: "grandparent",
				Args: []Term{Atom("alice"), Var("Who")},
			},
		}

		response := solveWithEvents(engine, query...)

		writeJSON(w, http.StatusOK, response)
	}
}

func SolveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()

		var request SolveRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON request body")
			return
		}

		if request.Program == "" {
			writeError(w, http.StatusBadRequest, "program is required")
			return
		}

		if request.Query == "" {
			writeError(w, http.StatusBadRequest, "query is required")
			return
		}

		clauses, err := ParseProgram(request.Program)
		if err != nil {
			writeError(w, http.StatusBadRequest, "program parse error: "+err.Error())
			return
		}

		goals, err := ParseQuery(request.Query)
		if err != nil {
			writeError(w, http.StatusBadRequest, "query parse error: "+err.Error())
			return
		}

		engine := Engine{
			Clauses: clauses,
		}

		response := solveWithEvents(engine, goals...)

		writeJSON(w, http.StatusOK, response)
	}
}

func solveWithEvents(engine Engine, goals ...Predicate) SolveResponse {
	var events []TraceEvent

	answers := engine.SolveWithTrace(func(event TraceEvent) {
		events = append(events, event)
	}, goals...)

	queryVars := queryVariables(goals)
	answerMaps := make([]map[string]string, len(answers))

	for i, answer := range answers {
		answerMaps[i] = answerBindings(answer, queryVars)
	}

	return SolveResponse{
		Events:  events,
		Answers: answerMaps,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: message,
	})
}
