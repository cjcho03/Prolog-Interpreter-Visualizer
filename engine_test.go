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

func TestSolveBuiltinIsAddsNumbers(t *testing.T) {
	query, err := ParseQuery("?- X is 2 + 3.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("5"))
}

func TestSolveBuiltinIsHonorsOperatorPrecedence(t *testing.T) {
	query, err := ParseQuery("?- X is 2 + 3 * 4.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("14"))
}

func TestSolveBuiltinIsHonorsParentheses(t *testing.T) {
	query, err := ParseQuery("?- X is (2 + 3) * 4.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("20"))
}

func TestSolveBuiltinIsUsesLeftAssociativity(t *testing.T) {
	query, err := ParseQuery("?- X is 10 - 3 - 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("5"))
}

func TestSolveBuiltinIsSupportsDivision(t *testing.T) {
	query, err := ParseQuery("?- X is 5 / 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("2.5"))
}

func TestSolveBuiltinIsPreservesFloatResult(t *testing.T) {
	query, err := ParseQuery("?- X is 2.5 + 0.5.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("3.0"))
}

func TestSolveBuiltinIsSupportsUnaryMinus(t *testing.T) {
	query, err := ParseQuery("?- X is -(2 + 3).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("-5"))
}

func TestSolveBuiltinIsUsesPreviouslyBoundVariable(t *testing.T) {
	query, err := ParseQuery("?- X = 5, Y is X * 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("5"))
	requireNumberBinding(t, answers[0], Var("Y"), Number("10"))
}

func TestSolveBuiltinIsInRuleBody(t *testing.T) {
	program := `
		double(X, Y) :- Y is X * 2.
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- double(6, Result).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{Clauses: clauses}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("Result"), Number("12"))
}

func TestSolveBuiltinIsFailsWithUnboundVariable(t *testing.T) {
	query, err := ParseQuery("?- X is Y + 1.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 0 {
		t.Fatalf("expected no answers for an unbound arithmetic variable, got %d", len(answers))
	}
}

func TestSolveBuiltinIsFailsOnDivisionByZero(t *testing.T) {
	query, err := ParseQuery("?- X is 5 / 0.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 0 {
		t.Fatalf("expected no answers for division by zero, got %d", len(answers))
	}
}

func TestUnificationDoesNotEvaluateArithmeticExpression(t *testing.T) {
	query, err := ParseQuery("?- X = 2 + 3.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	got := resolveTerm(Var("X"), answers[0])
	if got.String() != "2 + 3" {
		t.Fatalf("expected X to bind to the unevaluated expression, got %s", got.String())
	}
}

func requireNumberBinding(
	t *testing.T,
	sub Substitution,
	variable Var,
	want Number,
) {
	t.Helper()

	got := resolveTerm(variable, sub)
	number, ok := got.(Number)
	if !ok {
		t.Fatalf("expected %s to resolve to a number, got %T (%v)", variable, got, got)
	}

	if number != want {
		t.Fatalf("expected %s = %s, got %s", variable, want, number)
	}
}

func TestSolveNumericComparisonOperators(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		succeeds bool
	}{
		{name: "less true", query: "?- 3 < 5.", succeeds: true},
		{name: "less false", query: "?- 5 < 3.", succeeds: false},
		{name: "greater true", query: "?- 5 > 3.", succeeds: true},
		{name: "less equal true", query: "?- 5 =< 5.", succeeds: true},
		{name: "greater equal true", query: "?- 5 >= 5.", succeeds: true},
		{name: "arithmetic equal true", query: "?- 10 =:= 5 + 5.", succeeds: true},
		{name: "arithmetic equal false", query: "?- 10 =:= 5 + 4.", succeeds: false},
		{name: "arithmetic not equal true", query: "?- 4 =\\= 2 + 3.", succeeds: true},
		{name: "arithmetic not equal false", query: "?- 5 =\\= 2 + 3.", succeeds: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := ParseQuery(test.query)
			if err != nil {
				t.Fatal(err)
			}

			answers := (Engine{}).Solve(query...)
			if test.succeeds && len(answers) != 1 {
				t.Fatalf("expected comparison to succeed, got %d answers", len(answers))
			}
			if !test.succeeds && len(answers) != 0 {
				t.Fatalf("expected comparison to fail, got %d answers", len(answers))
			}
		})
	}
}

func TestSolveNumericComparisonEvaluatesBoundVariables(t *testing.T) {
	query, err := ParseQuery("?- X = 4, X + 1 =:= 5.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected comparison to succeed, got %d answers", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("4"))
}

func TestSolveNumericComparisonFailsWithUnboundVariable(t *testing.T) {
	query, err := ParseQuery("?- X > 3.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 0 {
		t.Fatalf("expected unbound arithmetic comparison to fail, got %d answers", len(answers))
	}
}

func TestSolveNumericComparisonInRuleBody(t *testing.T) {
	clauses, err := ParseProgram("adult(Age) :- Age >= 18.")
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- adult(21).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{Clauses: clauses}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected adult(21) to succeed, got %d answers", len(answers))
	}

	query, err = ParseQuery("?- adult(17).")
	if err != nil {
		t.Fatal(err)
	}

	answers = (Engine{Clauses: clauses}).Solve(query...)
	if len(answers) != 0 {
		t.Fatalf("expected adult(17) to fail, got %d answers", len(answers))
	}
}

func TestSolveBuiltinIsSupportsMod(t *testing.T) {
	query, err := ParseQuery("?- X is 10 mod 3.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("1"))
}

func TestSolveBuiltinIsRejectsFloatMod(t *testing.T) {
	query, err := ParseQuery("?- X is 5.5 mod 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 0 {
		t.Fatalf("expected float mod to fail, got %d answers", len(answers))
	}
}

func TestSolveBuiltinIsSupportsPower(t *testing.T) {
	query, err := ParseQuery("?- X is 2 ** 4.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("16"))
}

func TestSolveBuiltinIsPowerIsRightAssociative(t *testing.T) {
	query, err := ParseQuery("?- X is 2 ** 3 ** 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("512"))
}

func TestSolveBuiltinIsPowerBindsTighterThanUnaryMinus(t *testing.T) {
	query, err := ParseQuery("?- X is -2 ** 2.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{}).Solve(query...)
	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireNumberBinding(t, answers[0], Var("X"), Number("-4"))
}

func TestSolveBuiltinIsSupportsArithmeticFunctions(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  Number
	}{
		{name: "abs", query: "?- X is abs(-7).", want: Number("7")},
		{name: "min", query: "?- X is min(4, 9).", want: Number("4")},
		{name: "max", query: "?- X is max(4, 9).", want: Number("9")},
		{name: "nested", query: "?- X is max(abs(-7), min(4, 9)).", want: Number("7")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := ParseQuery(test.query)
			if err != nil {
				t.Fatal(err)
			}

			answers := (Engine{}).Solve(query...)
			if len(answers) != 1 {
				t.Fatalf("expected 1 answer, got %d", len(answers))
			}

			requireNumberBinding(t, answers[0], Var("X"), test.want)
		})
	}
}
