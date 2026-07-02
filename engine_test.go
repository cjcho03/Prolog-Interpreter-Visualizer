// engine_test.go
package prolog

import "testing"

func TestSolveSingleGoal(t *testing.T) {
	engine := Engine{
		Facts: []Predicate{
			{Name: "parent", Args: []Term{Atom("alice"), Atom("bob")}},
			{Name: "parent", Args: []Term{Atom("alice"), Atom("carol")}},
			{Name: "parent", Args: []Term{Atom("bob"), Atom("diana")}},
		},
	}

	answers := engine.Solve(
		Predicate{
			Name: "parent",
			Args: []Term{Atom("alice"), Var("Who")},
		},
	)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	if answers[0][Var("Who")] != Atom("bob") {
		t.Fatalf("expected first answer Who = bob, got %v", answers[0][Var("Who")])
	}

	if answers[1][Var("Who")] != Atom("carol") {
		t.Fatalf("expected second answer Who = carol, got %v", answers[1][Var("Who")])
	}
}

func TestSolveTwoGoals(t *testing.T) {
	engine := Engine{
		Facts: []Predicate{
			{Name: "parent", Args: []Term{Atom("alice"), Atom("bob")}},
			{Name: "parent", Args: []Term{Atom("alice"), Atom("carol")}},
			{Name: "parent", Args: []Term{Atom("bob"), Atom("diana")}},
			{Name: "parent", Args: []Term{Atom("carol"), Atom("eli")}},
		},
	}

	// parent(alice, X), parent(X, Y).
	answers := engine.Solve(
		Predicate{
			Name: "parent",
			Args: []Term{Atom("alice"), Var("X")},
		},
		Predicate{
			Name: "parent",
			Args: []Term{Var("X"), Var("Y")},
		},
	)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	if answers[0][Var("X")] != Atom("bob") ||
		answers[0][Var("Y")] != Atom("diana") {
		t.Fatalf("expected first answer X=bob, Y=diana; got %v", answers[0])
	}

	if answers[1][Var("X")] != Atom("carol") ||
		answers[1][Var("Y")] != Atom("eli") {
		t.Fatalf("expected second answer X=carol, Y=eli; got %v", answers[1])
	}
}

func TestSolveNoMatches(t *testing.T) {
	engine := Engine{
		Facts: []Predicate{
			{Name: "parent", Args: []Term{Atom("alice"), Atom("bob")}},
		},
	}

	answers := engine.Solve(
		Predicate{
			Name: "parent",
			Args: []Term{Atom("diana"), Var("Who")},
		},
	)

	if len(answers) != 0 {
		t.Fatalf("expected no answers, got %v", answers)
	}
}

func TestSolveBacktracksAfterFailure(t *testing.T) {
	engine := Engine{
		Facts: []Predicate{
			{Name: "parent", Args: []Term{Atom("alice"), Atom("bob")}},
			{Name: "parent", Args: []Term{Atom("alice"), Atom("carol")}},
			{Name: "parent", Args: []Term{Atom("carol"), Atom("eli")}},
		},
	}

	// The engine first tries X = bob, but bob has no child.
	// It must backtrack and try X = carol.
	answers := engine.Solve(
		Predicate{
			Name: "parent",
			Args: []Term{Atom("alice"), Var("X")},
		},
		Predicate{
			Name: "parent",
			Args: []Term{Var("X"), Var("Y")},
		},
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer after backtracking, got %d", len(answers))
	}

	if answers[0][Var("X")] != Atom("carol") {
		t.Fatalf("expected X = carol, got %v", answers[0][Var("X")])
	}

	if answers[0][Var("Y")] != Atom("eli") {
		t.Fatalf("expected Y = eli, got %v", answers[0][Var("Y")])
	}
}
