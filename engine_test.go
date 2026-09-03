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

func TestSolveFactWithCompoundTerm(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p(
				"likes",
				Atom("alice"),
				Compound{Name: "food", Args: []Term{Atom("pizza")}},
			)),
			Fact(p(
				"likes",
				Atom("bob"),
				Compound{Name: "food", Args: []Term{Atom("sushi")}},
			)),
		},
	}

	answers := engine.Solve(
		p("likes", Atom("alice"), Var("What")),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	got := resolveTerm(Var("What"), answers[0])

	if got.String() != "food(pizza)" {
		t.Fatalf("expected What = food(pizza), got %s", got.String())
	}
}

func TestSolveMatchesInsideCompoundTerm(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p(
				"likes",
				Atom("alice"),
				Compound{Name: "food", Args: []Term{Atom("pizza")}},
			)),
			Fact(p(
				"likes",
				Atom("bob"),
				Compound{Name: "food", Args: []Term{Atom("sushi")}},
			)),
		},
	}

	answers := engine.Solve(
		p(
			"likes",
			Var("Person"),
			Compound{Name: "food", Args: []Term{Atom("pizza")}},
		),
	)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Person"), Atom("alice"))
}

func TestSolveFactWithList(t *testing.T) {
	program := `
		likes(alice, [pizza, sushi]).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- likes(alice, Foods).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	got := resolveTerm(Var("Foods"), answers[0])

	if got.String() != "[pizza, sushi]" {
		t.Fatalf("expected Foods = [pizza, sushi], got %s", got.String())
	}
}

func TestSolveListHeadAndTail(t *testing.T) {
	program := `
		items([pizza, sushi, ramen]).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- items([Head | Tail]).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Head"), Atom("pizza"))

	tail := resolveTerm(Var("Tail"), answers[0])

	if tail.String() != "[sushi, ramen]" {
		t.Fatalf("expected Tail = [sushi, ramen], got %s", tail.String())
	}
}

func TestSolveBuiltinTrue(t *testing.T) {
	query, err := ParseQuery("?- true.")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{}
	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected true/0 to produce 1 answer, got %d", len(answers))
	}
}

func TestSolveBuiltinFail(t *testing.T) {
	query, err := ParseQuery("?- fail.")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{}
	answers := engine.Solve(query...)

	if len(answers) != 0 {
		t.Fatalf("expected fail/0 to produce no answers, got %d", len(answers))
	}
}

func TestSolveBuiltinTrueInRuleBody(t *testing.T) {
	program := `
		item(pizza).
		item(sushi).

		available(X) :- item(X), true.
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- available(Item).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}
	answers := engine.Solve(query...)

	if len(answers) != 2 {
		t.Fatalf("expected 2 answers, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Item"), Atom("pizza"))
	requireAtomBinding(t, answers[1], Var("Item"), Atom("sushi"))
}

func TestSolveBuiltinFailBacktracksToAnotherClause(t *testing.T) {
	program := `
		choice(first) :- fail.
		choice(second).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- choice(Value).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}
	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer after fail/0 backtracking, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("Value"), Atom("second"))
}

func TestBuiltinLookupUsesPredicateArity(t *testing.T) {
	engine := Engine{
		Clauses: []Clause{
			Fact(p("true", Atom("value"))),
		},
	}

	answers := engine.Solve(p("true", Var("X")))

	if len(answers) != 1 {
		t.Fatalf("expected user-defined true/1 to be treated as a normal predicate, got %d answers", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("value"))
}

func TestSolveBuiltinUnifyBindsVariable(t *testing.T) {
	query, err := ParseQuery("?- X = alice.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("alice"))
}

func TestSolveBuiltinUnifyFailsForDifferentAtoms(t *testing.T) {
	query, err := ParseQuery("?- alice = bob.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)

	if len(answers) != 0 {
		t.Fatalf("expected no answers, got %d", len(answers))
	}
}

func TestSolveBuiltinUnifyVariableChain(t *testing.T) {
	query, err := ParseQuery("?- X = Y, Y = alice.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("alice"))
	requireAtomBinding(t, answers[0], Var("Y"), Atom("alice"))
}

func TestSolveBuiltinUnifyCompoundTerms(t *testing.T) {
	query, err := ParseQuery("?- food(X) = food(pizza).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("pizza"))
}

func TestSolveBuiltinUnifyLists(t *testing.T) {
	query, err := ParseQuery("?- [X, bob] = [alice, Y].")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireAtomBinding(t, answers[0], Var("X"), Atom("alice"))
	requireAtomBinding(t, answers[0], Var("Y"), Atom("bob"))
}

func TestSolveBuiltinUnifyInRuleBody(t *testing.T) {
	clauses, err := ParseProgram("same(X, Y) :- X = Y.")
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- same(alice, alice).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{Clauses: clauses}).Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected matching terms to succeed, got %d answers", len(answers))
	}

	query, err = ParseQuery("?- same(alice, bob).")
	if err != nil {
		t.Fatal(err)
	}

	answers = (Engine{Clauses: clauses}).Solve(query...)

	if len(answers) != 0 {
		t.Fatalf("expected different terms to fail, got %d answers", len(answers))
	}
}
