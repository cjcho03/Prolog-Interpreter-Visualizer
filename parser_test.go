package prolog

import "testing"

func TestParseFact(t *testing.T) {
	clauses, err := ParseProgram("parent(alice, bob).")
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}

	if !clauses[0].IsFact() {
		t.Fatal("expected parsed clause to be a fact")
	}

	if clauses[0].Head.String() != "parent(alice, bob)" {
		t.Fatalf("unexpected fact head: %s", clauses[0].Head.String())
	}
}

func TestParseRule(t *testing.T) {
	src := `
		grandparent(X, Z) :-
			parent(X, Y),
			parent(Y, Z).
	`

	clauses, err := ParseProgram(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}

	if clauses[0].IsFact() {
		t.Fatal("expected parsed clause to be a rule")
	}

	got := clauses[0].String()
	want := "grandparent(X, Z) :- parent(X, Y), parent(Y, Z)"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseMultipleClauses(t *testing.T) {
	src := `
		parent(alice, bob).
		parent(bob, diana).

		grandparent(X, Z) :-
			parent(X, Y),
			parent(Y, Z).
	`

	clauses, err := ParseProgram(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 3 {
		t.Fatalf("expected 3 clauses, got %d", len(clauses))
	}

	if !clauses[0].IsFact() {
		t.Fatal("expected first clause to be a fact")
	}

	if !clauses[1].IsFact() {
		t.Fatal("expected second clause to be a fact")
	}

	if clauses[2].IsFact() {
		t.Fatal("expected third clause to be a rule")
	}
}

func TestParseQuery(t *testing.T) {
	goals, err := ParseQuery("?- grandparent(alice, Who).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	got := goals[0].String()
	want := "grandparent(alice, Who)"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseMultipleGoalQuery(t *testing.T) {
	goals, err := ParseQuery("?- parent(alice, X), parent(X, Y).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 2 {
		t.Fatalf("expected 2 query goals, got %d", len(goals))
	}

	if goals[0].String() != "parent(alice, X)" {
		t.Fatalf("unexpected first goal: %s", goals[0].String())
	}

	if goals[1].String() != "parent(X, Y)" {
		t.Fatalf("unexpected second goal: %s", goals[1].String())
	}
}

func TestParseProgramThenSolve(t *testing.T) {
	program := `
		parent(alice, bob).
		parent(alice, carol).
		parent(bob, diana).
		parent(carol, eli).

		grandparent(X, Z) :-
			parent(X, Y),
			parent(Y, Z).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- grandparent(alice, Who).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	requireParserAtomBinding(t, answers[0], Var("Who"), Atom("diana"))
	requireParserAtomBinding(t, answers[1], Var("Who"), Atom("eli"))
}

func TestParseSupportsComments(t *testing.T) {
	src := `
		% family facts
		parent(alice, bob). % inline comment

		% family rule
		grandparent(X, Z) :-
			parent(X, Y),
			parent(Y, Z).
	`

	clauses, err := ParseProgram(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 2 {
		t.Fatalf("expected 2 clauses, got %d", len(clauses))
	}
}

func TestParseRejectsQueryInsideProgram(t *testing.T) {
	_, err := ParseProgram("?- parent(alice, Who).")
	if err == nil {
		t.Fatal("expected ParseProgram to reject query syntax")
	}
}

func TestParseSupportsAnonymousVariable(t *testing.T) {
	goals, err := ParseQuery("?- parent(alice, _).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	got := goals[0].String()
	want := "parent(alice, _)"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseAnonymousVariablesAreDistinct(t *testing.T) {
	goals, err := ParseQuery("?- same(_, _).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	first, ok := goals[0].Args[0].(Var)
	if !ok {
		t.Fatalf("expected first anonymous argument to be a Var")
	}

	second, ok := goals[0].Args[1].(Var)
	if !ok {
		t.Fatalf("expected second anonymous argument to be a Var")
	}

	if first == second {
		t.Fatalf("expected each anonymous variable to be fresh, got %s", first)
	}
}

func TestParseProgramThenSolveWithAnonymousVariable(t *testing.T) {
	program := `
		parent(alice, bob).
		parent(alice, carol).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- parent(alice, _).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}
}

func TestParseProgramThenSolveWithAnonymousVariableAndNamedVariable(t *testing.T) {
	program := `
		parent(alice, bob).
		parent(carol, eli).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- parent(_, Who).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	requireParserAtomBinding(t, answers[0], Var("Who"), Atom("bob"))
	requireParserAtomBinding(t, answers[1], Var("Who"), Atom("eli"))
}

func TestParseRejectsNestedCompoundTermsForNow(t *testing.T) {
	_, err := ParseQuery("?- likes(alice, food(pizza)).")
	if err == nil {
		t.Fatal("expected nested compound term to be rejected for now")
	}
}

func requireParserAtomBinding(
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

func TestParseNumberFact(t *testing.T) {
	clauses, err := ParseProgram("age(alice, 30).")
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}

	got := clauses[0].Head.String()
	want := "age(alice, 30)"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseNumberQuery(t *testing.T) {
	goals, err := ParseQuery("?- age(alice, 30).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	got := goals[0].String()
	want := "age(alice, 30)"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseProgramThenSolveWithNumber(t *testing.T) {
	program := `
		age(alice, 30).
		age(bob, 24).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- age(alice, Age).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireParserNumberBinding(t, answers[0], Var("Age"), Number("30"))
}

func requireParserNumberBinding(
	t *testing.T,
	answer Substitution,
	variable Var,
	want Number,
) {
	t.Helper()

	got := dereference(variable, answer)

	number, ok := got.(Number)
	if !ok {
		t.Fatalf("expected %s to resolve to a number, got %v", variable, got)
	}

	if number != want {
		t.Fatalf("expected %s = %s, got %s", variable, want, number)
	}
}
