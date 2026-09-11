package prolog

import "testing"

func TestLexStructuralNotUnifiableOperator(t *testing.T) {
	tokens, err := lex(`?- a \= b.`)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, tok := range tokens {
		if tok.typ == tokenNotUnifiable {
			found = true
			if tok.literal != `\=` {
				t.Fatalf("expected \\= literal, got %q", tok.literal)
			}
		}
	}

	if !found {
		t.Fatal("expected lexer to emit tokenNotUnifiable")
	}
}

func TestParseStructuralNotUnifiableGoal(t *testing.T) {
	goals, err := ParseQuery(`?- a \= b.`)
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(goals))
	}

	if goals[0].Name != `\=` || len(goals[0].Args) != 2 {
		t.Fatalf("expected \\=/2, got %s/%d", goals[0].Name, len(goals[0].Args))
	}

	if got := goals[0].String(); got != `a \= b` {
		t.Fatalf("expected %q, got %q", `a \= b`, got)
	}
}

func TestStructuralNotUnifiableSucceedsForDifferentAtoms(t *testing.T) {
	goals, err := ParseQuery(`?- a \= b.`)
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected a \\= b to succeed, got %d answers", len(answers))
	}
}

func TestStructuralNotUnifiableFailsForEqualAtoms(t *testing.T) {
	goals, err := ParseQuery(`?- a \= a.`)
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(goals...)
	if len(answers) != 0 {
		t.Fatalf("expected a \\= a to fail, got %d answers", len(answers))
	}
}

func TestStructuralNotUnifiableFailsWhenTermsCouldUnify(t *testing.T) {
	queries := []string{
		`?- X \= a.`,
		`?- pair(X, b) \= pair(a, b).`,
		`?- [X, b] \= [a, b].`,
	}

	for _, query := range queries {
		goals, err := ParseQuery(query)
		if err != nil {
			t.Fatalf("parse %q: %v", query, err)
		}

		answers := (Engine{}).Solve(goals...)
		if len(answers) != 0 {
			t.Fatalf("expected %q to fail because the terms are unifiable", query)
		}
	}
}

func TestStructuralNotUnifiableUsesExistingBindings(t *testing.T) {
	goals, err := ParseQuery(`?- X = a, X \= b.`)
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected bound X = a to differ from b, got %d answers", len(answers))
	}

	if got := dereference(Var("X"), answers[0]); got != Atom("a") {
		t.Fatalf("expected X = a, got %v", got)
	}
}

func TestStructuralNotUnifiableDoesNotLeakProbeBindings(t *testing.T) {
	goals, err := ParseQuery(`?- pair(X, a) \= pair(b, c), X = d.`)
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected query to succeed, got %d answers", len(answers))
	}

	if got := dereference(Var("X"), answers[0]); got != Atom("d") {
		t.Fatalf("expected probe binding not to leak and X = d, got %v", got)
	}
}

func TestStructuralNotUnifiableEmitsBuiltinTrace(t *testing.T) {
	goals, err := ParseQuery(`?- a \= b.`)
	if err != nil {
		t.Fatal(err)
	}

	var events []TraceEvent
	answers := (Engine{}).SolveWithTrace(func(event TraceEvent) {
		events = append(events, event)
	}, goals...)

	if len(answers) != 1 {
		t.Fatalf("expected query to succeed, got %d answers", len(answers))
	}

	for _, event := range events {
		if event.Type == EventBuiltin && event.Goal == `a \= b` {
			return
		}
	}

	t.Fatal("expected trace to include builtin event for \\=/2")
}
