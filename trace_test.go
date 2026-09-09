package prolog

import (
	"strings"
	"testing"
)

func collectTrace(
	engine Engine,
	goals ...Predicate,
) ([]TraceEvent, []Substitution) {
	var events []TraceEvent

	answers := engine.SolveWithTrace(func(event TraceEvent) {
		events = append(events, event)
	}, goals...)

	return events, answers
}

func findFirstEvent(
	events []TraceEvent,
	eventType TraceEventType,
) (TraceEvent, bool) {
	for _, event := range events {
		if event.Type == eventType {
			return event, true
		}
	}

	return TraceEvent{}, false
}

func TestSolveTraceUsesClauseAttempts(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
		},
	}

	events, _ := collectTrace(
		engine,
		p("parent", Atom("alice"), Var("Who")),
	)

	event, found := findFirstEvent(events, EventTryClause)

	if !found {
		t.Fatal("expected trace to include a try_clause event")
	}

	if event.Goal != "parent(alice, Who)" {
		t.Fatalf("expected goal parent(alice, Who), got %q", event.Goal)
	}

	if event.Clause != "parent(alice, bob)" {
		t.Fatalf("expected parent clause, got %q", event.Clause)
	}
}

func TestSolveTraceEmitsRuleExpansion(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("bob"), Atom("diana"))),

			Rule(
				p("grandparent", Var("X"), Var("Z")),
				p("parent", Var("X"), Var("Y")),
				p("parent", Var("Y"), Var("Z")),
			),
		},
	}

	events, answers := collectTrace(
		engine,
		p("grandparent", Atom("alice"), Var("Who")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	event, found := findFirstEvent(events, EventRuleExpanded)

	if !found {
		t.Fatal("expected trace to include a rule_expanded event")
	}

	if event.Goal != "grandparent(alice, Who)" {
		t.Fatalf("expected grandparent goal, got %q", event.Goal)
	}

	if !strings.Contains(event.Clause, "grandparent") {
		t.Fatalf("expected rule clause in trace, got %q", event.Clause)
	}

	if len(event.ExpandedGoals) != 2 {
		t.Fatalf(
			"expected rule expansion to contain 2 goals, got %d",
			len(event.ExpandedGoals),
		)
	}

	if !strings.HasPrefix(event.ExpandedGoals[0], "parent(alice,") {
		t.Fatalf(
			"expected first expanded goal to begin with parent(alice, ...), got %q",
			event.ExpandedGoals[0],
		)
	}
}

func TestSolveTraceEmitsBacktrackingForDeadEndRuleBranch(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("alice"), Atom("carol"))),
			Fact(p("parent", Atom("carol"), Atom("eli"))),

			Rule(
				p("grandparent", Var("X"), Var("Z")),
				p("parent", Var("X"), Var("Y")),
				p("parent", Var("Y"), Var("Z")),
			),
		},
	}

	events, _ := collectTrace(
		engine,
		p("grandparent", Atom("alice"), Var("Who")),
	)

	foundBacktrack := false

	for _, event := range events {
		if event.Type != EventBacktrack {
			continue
		}

		if strings.Contains(event.Description, "Backtracking") {
			foundBacktrack = true
			break
		}
	}

	if !foundBacktrack {
		t.Fatal("expected a dead-end rule branch to emit a backtrack event")
	}
}

func TestSolveTraceEmitsBuiltinEvent(t *testing.T) {
	engine := Engine{}

	events, answers := collectTrace(engine, p("true"))

	if len(answers) != 1 {
		t.Fatalf("expected true/0 to produce 1 answer, got %d", len(answers))
	}

	event, found := findFirstEvent(events, EventBuiltin)
	if !found {
		t.Fatal("expected trace to include a builtin event")
	}

	if event.Goal != "true()" {
		t.Fatalf("expected builtin goal true(), got %q", event.Goal)
	}
}

func TestSolveTraceShowsBuiltinFailure(t *testing.T) {
	engine := Engine{}

	events, answers := collectTrace(engine, p("fail"))

	if len(answers) != 0 {
		t.Fatalf("expected fail/0 to produce no answers, got %d", len(answers))
	}

	failed, found := findFirstEvent(events, EventFailed)
	if !found {
		t.Fatal("expected trace to include a failed event")
	}

	if !strings.Contains(failed.Description, "Built-in predicate failed") {
		t.Fatalf("expected built-in failure description, got %q", failed.Description)
	}
}

func TestSolveTraceShowsUnificationBinding(t *testing.T) {
	query, err := ParseQuery("?- X = alice.")
	if err != nil {
		t.Fatal(err)
	}

	events, answers := collectTrace(Engine{}, query...)

	if len(answers) != 1 {
		t.Fatalf("expected =/2 to produce 1 answer, got %d", len(answers))
	}

	builtin, found := findFirstEvent(events, EventBuiltin)
	if !found {
		t.Fatal("expected trace to include a builtin event")
	}

	if builtin.Goal != "X = alice" {
		t.Fatalf("expected builtin goal X = alice, got %q", builtin.Goal)
	}

	solution, found := findFirstEvent(events, EventSolution)
	if !found {
		t.Fatal("expected trace to include a solution event")
	}

	if solution.Bindings["X"] != "alice" {
		t.Fatalf("expected solution trace to show X = alice, got %v", solution.Bindings)
	}
}

func TestSolveTraceShowsArithmeticBuiltinBinding(t *testing.T) {
	query, err := ParseQuery("?- X is 2 + 3 * 4.")
	if err != nil {
		t.Fatal(err)
	}

	var events []TraceEvent
	answers := (Engine{}).SolveWithTrace(func(event TraceEvent) {
		events = append(events, event)
	}, query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	builtin, found := findFirstEvent(events, EventBuiltin)
	if !found {
		t.Fatal("expected trace to include a builtin event")
	}

	if builtin.Goal != "X is 2 + 3 * 4" {
		t.Fatalf("expected arithmetic builtin goal, got %q", builtin.Goal)
	}

	solution, found := findFirstEvent(events, EventSolution)
	if !found {
		t.Fatal("expected trace to include a solution event")
	}

	if solution.Bindings["X"] != "14" {
		t.Fatalf("expected solution trace to show X = 14, got %q", solution.Bindings["X"])
	}
}

func TestSolveTraceShowsNumericComparisonBuiltin(t *testing.T) {
	query, err := ParseQuery("?- 2 + 3 =:= 5.")
	if err != nil {
		t.Fatal(err)
	}

	events, answers := collectTrace(Engine{}, query...)
	if len(answers) != 1 {
		t.Fatalf("expected comparison to succeed, got %d answers", len(answers))
	}

	builtin, found := findFirstEvent(events, EventBuiltin)
	if !found {
		t.Fatal("expected trace to include a builtin event")
	}

	if builtin.Goal != "2 + 3 =:= 5" {
		t.Fatalf("expected comparison builtin goal, got %q", builtin.Goal)
	}
}
