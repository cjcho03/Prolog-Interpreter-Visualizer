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
