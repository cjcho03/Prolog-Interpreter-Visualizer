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

func TestParseUnificationQuery(t *testing.T) {
	goals, err := ParseQuery("?- X = alice.")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	if goals[0].Name != "=" || len(goals[0].Args) != 2 {
		t.Fatalf("expected =/2, got %s/%d", goals[0].Name, len(goals[0].Args))
	}

	if got := goals[0].String(); got != "X = alice" {
		t.Fatalf("expected %q, got %q", "X = alice", got)
	}
}

func TestParseUnificationWithCompoundTerms(t *testing.T) {
	goals, err := ParseQuery("?- food(X) = food(pizza).")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "food(X) = food(pizza)" {
		t.Fatalf("expected %q, got %q", "food(X) = food(pizza)", got)
	}
}

func TestParseUnificationWithLists(t *testing.T) {
	goals, err := ParseQuery("?- [X, bob] = [alice, Y].")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "[X, bob] = [alice, Y]" {
		t.Fatalf("expected %q, got %q", "[X, bob] = [alice, Y]", got)
	}
}

func TestParseUnificationInRuleBody(t *testing.T) {
	clauses, err := ParseProgram("same(X, Y) :- X = Y.")
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}

	if got := clauses[0].String(); got != "same(X, Y) :- X = Y" {
		t.Fatalf("expected %q, got %q", "same(X, Y) :- X = Y", got)
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

func TestParseCompoundTerm(t *testing.T) {
	goals, err := ParseQuery("?- likes(alice, food(pizza)).")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	got := goals[0].String()
	want := "likes(alice, food(pizza))"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseNestedCompoundTerm(t *testing.T) {
	goals, err := ParseQuery("?- likes(alice, meal(food(pizza), drink(water))).")
	if err != nil {
		t.Fatal(err)
	}

	got := goals[0].String()
	want := "likes(alice, meal(food(pizza), drink(water)))"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseProgramThenSolveWithCompoundTerm(t *testing.T) {
	program := `
		likes(alice, food(pizza)).
		likes(bob, food(sushi)).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- likes(alice, What).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	got := dereference(Var("What"), answers[0])

	compound, ok := got.(Compound)
	if !ok {
		t.Fatalf("expected What to resolve to a compound term, got %v", got)
	}

	if compound.String() != "food(pizza)" {
		t.Fatalf("expected What = food(pizza), got %s", compound.String())
	}
}

func TestParseProgramThenSolveWithVariableInsideCompoundTerm(t *testing.T) {
	program := `
		likes(alice, food(pizza)).
		likes(bob, food(sushi)).
	`

	clauses, err := ParseProgram(program)
	if err != nil {
		t.Fatal(err)
	}

	query, err := ParseQuery("?- likes(Person, food(pizza)).")
	if err != nil {
		t.Fatal(err)
	}

	engine := Engine{Clauses: clauses}

	answers := engine.Solve(query...)

	if len(answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(answers))
	}

	requireParserAtomBinding(t, answers[0], Var("Person"), Atom("alice"))
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

func TestParseEmptyList(t *testing.T) {
	goals, err := ParseQuery("?- items([]).")
	if err != nil {
		t.Fatal(err)
	}

	got := goals[0].String()
	want := "items([])"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseList(t *testing.T) {
	goals, err := ParseQuery("?- items([pizza, sushi, ramen]).")
	if err != nil {
		t.Fatal(err)
	}

	got := goals[0].String()
	want := "items([pizza, sushi, ramen])"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseListWithTail(t *testing.T) {
	goals, err := ParseQuery("?- items([Head | Tail]).")
	if err != nil {
		t.Fatal(err)
	}

	got := goals[0].String()
	want := "items([Head | Tail])"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseListWithElementsAndTail(t *testing.T) {
	goals, err := ParseQuery("?- items([pizza, sushi | Rest]).")
	if err != nil {
		t.Fatal(err)
	}

	got := goals[0].String()
	want := "items([pizza, sushi | Rest])"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
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

func TestParseIsArithmeticQuery(t *testing.T) {
	goals, err := ParseQuery("?- X is 2 + 3 * 4.")
	if err != nil {
		t.Fatal(err)
	}

	if len(goals) != 1 {
		t.Fatalf("expected 1 query goal, got %d", len(goals))
	}

	if goals[0].Name != "is" || len(goals[0].Args) != 2 {
		t.Fatalf("expected is/2, got %s/%d", goals[0].Name, len(goals[0].Args))
	}

	if got := goals[0].String(); got != "X is 2 + 3 * 4" {
		t.Fatalf("expected %q, got %q", "X is 2 + 3 * 4", got)
	}
}

func TestParseArithmeticParenthesesPreserveGrouping(t *testing.T) {
	goals, err := ParseQuery("?- X is (2 + 3) * 4.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is (2 + 3) * 4" {
		t.Fatalf("expected %q, got %q", "X is (2 + 3) * 4", got)
	}
}

func TestParseArithmeticIsLeftAssociative(t *testing.T) {
	goals, err := ParseQuery("?- X is 10 - 3 - 2.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is 10 - 3 - 2" {
		t.Fatalf("expected %q, got %q", "X is 10 - 3 - 2", got)
	}
}

func TestParseArithmeticRightGroupingIsPrinted(t *testing.T) {
	goals, err := ParseQuery("?- X is 10 - (3 - 2).")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is 10 - (3 - 2)" {
		t.Fatalf("expected %q, got %q", "X is 10 - (3 - 2)", got)
	}
}

func TestParseArithmeticSupportsDecimalNumbers(t *testing.T) {
	goals, err := ParseQuery("?- X is 2.5 + 0.5.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is 2.5 + 0.5" {
		t.Fatalf("expected %q, got %q", "X is 2.5 + 0.5", got)
	}
}

func TestParseArithmeticSupportsUnaryMinus(t *testing.T) {
	goals, err := ParseQuery("?- X is -(2 + 3).")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is -(2 + 3)" {
		t.Fatalf("expected %q, got %q", "X is -(2 + 3)", got)
	}
}

func TestParseUnificationCanBindArithmeticExpression(t *testing.T) {
	goals, err := ParseQuery("?- X = 2 + 3 * 4.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X = 2 + 3 * 4" {
		t.Fatalf("expected %q, got %q", "X = 2 + 3 * 4", got)
	}
}

func TestParseIsInRuleBody(t *testing.T) {
	clauses, err := ParseProgram("double(X, Y) :- Y is X * 2.")
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}

	if got := clauses[0].String(); got != "double(X, Y) :- Y is X * 2" {
		t.Fatalf("expected %q, got %q", "double(X, Y) :- Y is X * 2", got)
	}
}

func TestParseNumericComparisonOperators(t *testing.T) {
	tests := []struct {
		query string
		want  string
		name  string
	}{
		{query: "?- 3 < 5.", want: "3 < 5", name: "<"},
		{query: "?- 5 > 3.", want: "5 > 3", name: ">"},
		{query: "?- 3 =< 5.", want: "3 =< 5", name: "=<"},
		{query: "?- 5 >= 5.", want: "5 >= 5", name: ">="},
		{query: "?- 10 =:= 5 + 5.", want: "10 =:= 5 + 5", name: "=:="},
		{query: "?- 4 =\\= 2 + 3.", want: "4 =\\= 2 + 3", name: "=\\="},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			goals, err := ParseQuery(test.query)
			if err != nil {
				t.Fatal(err)
			}

			if len(goals) != 1 {
				t.Fatalf("expected 1 query goal, got %d", len(goals))
			}

			if goals[0].Name != test.name || len(goals[0].Args) != 2 {
				t.Fatalf("expected %s/2, got %s/%d", test.name, goals[0].Name, len(goals[0].Args))
			}

			if got := goals[0].String(); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestParseNumericComparisonInRuleBody(t *testing.T) {
	clauses, err := ParseProgram("adult(Age) :- Age >= 18.")
	if err != nil {
		t.Fatal(err)
	}

	if got := clauses[0].String(); got != "adult(Age) :- Age >= 18" {
		t.Fatalf("expected %q, got %q", "adult(Age) :- Age >= 18", got)
	}
}

func TestParseArithmeticSupportsMod(t *testing.T) {
	goals, err := ParseQuery("?- X is 10 mod 3.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is 10 mod 3" {
		t.Fatalf("expected %q, got %q", "X is 10 mod 3", got)
	}
}

func TestParseArithmeticSupportsPower(t *testing.T) {
	goals, err := ParseQuery("?- X is 2 ** 3 ** 2.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is 2 ** 3 ** 2" {
		t.Fatalf("expected %q, got %q", "X is 2 ** 3 ** 2", got)
	}
}

func TestParseArithmeticPowerPrecedence(t *testing.T) {
	goals, err := ParseQuery("?- X is -2 ** 2 + 3 * 4.")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is -2 ** 2 + 3 * 4" {
		t.Fatalf("expected %q, got %q", "X is -2 ** 2 + 3 * 4", got)
	}
}

func TestParseArithmeticSupportsFunctions(t *testing.T) {
	goals, err := ParseQuery("?- X is max(abs(-7), min(4, 9)).")
	if err != nil {
		t.Fatal(err)
	}

	if got := goals[0].String(); got != "X is max(abs(-7), min(4, 9))" {
		t.Fatalf("expected %q, got %q", "X is max(abs(-7), min(4, 9))", got)
	}
}
