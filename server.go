package prolog

import (
	"encoding/json"
	"net/http"
	"strings"
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

type HealthResponse struct {
	Status string `json:"status"`
}

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
	}
}

// CORSMiddleware allows the browser visualizer to call the API when the two are
// deployed on different origins. Pass exact frontend origins via ALLOWED_ORIGINS.
// Local Vite development normally uses the same-origin /api proxy and therefore
// does not require a CORS origin.
func CORSMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAny := false

	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAny = true
			continue
		}
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))

		if origin != "" {
			_, explicitlyAllowed := allowed[origin]
			if !allowAny && !explicitlyAllowed {
				writeError(w, http.StatusForbidden, "origin not allowed")
				return
			}

			if allowAny {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
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
