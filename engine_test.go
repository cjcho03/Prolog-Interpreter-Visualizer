package prolog

import "testing"

func p(name string, args ...Term) Predicate {
	return Predicate{
		Name: name,
		Args: args,
	}
}

func requireAtomBinding(
	t *testing.T,
	answer Substitution,
	variable Var,
	want Atom,
) {
	t.Helper()

	got := dereference(variable, answer)

	atom, ok := got.(Atom)
	if !ok {
		t.Fatalf("expected %s to resolve to an atom, got %v", variable, got)
	}

	if atom != want {
		t.Fatalf("expected %s = %s, got %s", variable, want, atom)
	}
}

func TestSolveSingleFactGoal(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("alice"), Atom("carol"))),
			Fact(p("parent", Atom("bob"), Atom("diana"))),
		},
	}

	answers := engine.Solve(
		p("parent", Atom("alice"), Var("Who")),
	)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Who"), Atom("bob"))
	requireAtomBinding(t, answers[1], Var("Who"), Atom("carol"))
}

func TestSolveMultipleFactGoals(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("alice"), Atom("carol"))),
			Fact(p("parent", Atom("bob"), Atom("diana"))),
			Fact(p("parent", Atom("carol"), Atom("eli"))),
		},
	}

	answers := engine.Solve(
		p("parent", Atom("alice"), Var("X")),
		p("parent", Var("X"), Var("Y")),
	)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("bob"))
	requireAtomBinding(t, answers[0], Var("Y"), Atom("diana"))

	requireAtomBinding(t, answers[1], Var("X"), Atom("carol"))
	requireAtomBinding(t, answers[1], Var("Y"), Atom("eli"))
}

func TestSolveNoMatches(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
		},
	}

	answers := engine.Solve(
		p("parent", Atom("diana"), Var("Who")),
	)

	if len(answers) != 0 {
		t.Fatalf("expected no answers, got %v", answers)
	}
}

func TestSolveBacktracksAcrossFacts(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("alice"), Atom("carol"))),
			Fact(p("parent", Atom("carol"), Atom("eli"))),
		},
	}

	answers := engine.Solve(
		p("parent", Atom("alice"), Var("X")),
		p("parent", Var("X"), Var("Y")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer after backtracking, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("carol"))
	requireAtomBinding(t, answers[0], Var("Y"), Atom("eli"))
}

func TestSolveRule(t *testing.T) {
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

	answers := engine.Solve(
		p("grandparent", Atom("alice"), Var("Who")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Who"), Atom("diana"))
}

func TestSolveRuleBodyBacktracks(t *testing.T) {
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

	answers := engine.Solve(
		p("grandparent", Atom("alice"), Var("Who")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Who"), Atom("eli"))
}

func TestSolveRecursiveRule(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("alice"), Atom("carol"))),
			Fact(p("parent", Atom("bob"), Atom("diana"))),
			Fact(p("parent", Atom("carol"), Atom("eli"))),

			Rule(
				p("ancestor", Var("X"), Var("Y")),
				p("parent", Var("X"), Var("Y")),
			),

			Rule(
				p("ancestor", Var("X"), Var("Y")),
				p("parent", Var("X"), Var("Z")),
				p("ancestor", Var("Z"), Var("Y")),
			),
		},
	}

	answers := engine.Solve(
		p("ancestor", Atom("alice"), Var("Who")),
	)

	if len(answers) != 4 {
		t.Fatalf("expected 4 answers, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Who"), Atom("bob"))
	requireAtomBinding(t, answers[1], Var("Who"), Atom("carol"))
	requireAtomBinding(t, answers[2], Var("Who"), Atom("diana"))
	requireAtomBinding(t, answers[3], Var("Who"), Atom("eli"))
}

func TestSolveStandardizesRuleVariablesApart(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("parent", Atom("alice"), Atom("bob"))),
			Fact(p("parent", Atom("carol"), Atom("eli"))),

			Rule(
				p("ancestor", Var("X"), Var("Y")),
				p("parent", Var("X"), Var("Y")),
			),
		},
	}

	answers := engine.Solve(
		p("ancestor", Atom("alice"), Var("First")),
		p("ancestor", Atom("carol"), Var("Second")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("First"), Atom("bob"))
	requireAtomBinding(t, answers[0], Var("Second"), Atom("eli"))
}

func TestSolveFactWithNumber(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("age", Atom("alice"), Number("30"))),
			Fact(p("age", Atom("bob"), Number("24"))),
		},
	}

	answers := engine.Solve(
		p("age", Atom("alice"), Var("Age")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	got := dereference(Var("Age"), answers[0])

	number, ok := got.(Number)
	if !ok {
		t.Fatalf("expected Age to resolve to a number, got %v", got)
	}

	if number != Number("30") {
		t.Fatalf("expected Age = 30, got %s", number)
	}
}
