package prolog

import "testing"

func TestFactCreatesClauseWithoutBody(t *testing.T) {
	head := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Atom("bob")},
	}

	clause := Fact(head)

	if clause.Head.String() != "parent(alice, bob)" {
		t.Fatalf("unexpected fact head: %s", clause.Head.String())
	}

	if len(clause.Body) != 0 {
		t.Fatalf("expected fact to have no body, got %v", clause.Body)
	}
}

func TestRuleCreatesClauseWithBody(t *testing.T) {
	head := Predicate{
		Name: "grandparent",
		Args: []Term{Var("X"), Var("Z")},
	}

	firstGoal := Predicate{
		Name: "parent",
		Args: []Term{Var("X"), Var("Y")},
	}

	secondGoal := Predicate{
		Name: "parent",
		Args: []Term{Var("Y"), Var("Z")},
	}

	clause := Rule(head, firstGoal, secondGoal)

	if clause.Head.String() != "grandparent(X, Z)" {
		t.Fatalf("unexpected rule head: %s", clause.Head.String())
	}

	if len(clause.Body) != 2 {
		t.Fatalf("expected 2 rule goals, got %d", len(clause.Body))
	}

	if clause.Body[0].String() != "parent(X, Y)" {
		t.Fatalf("unexpected first rule goal: %s", clause.Body[0].String())
	}

	if clause.Body[1].String() != "parent(Y, Z)" {
		t.Fatalf("unexpected second rule goal: %s", clause.Body[1].String())
	}
}

func TestStandardizeApartRenamesVariablesConsistently(t *testing.T) {
	clause := Rule(
		p("grandparent", Var("X"), Var("Z")),
		p("parent", Var("X"), Var("Y")),
		p("parent", Var("Y"), Var("Z")),
	)

	got := standardizeApart(clause, 7)

	if got.Head.String() != "grandparent($7_X, $7_Z)" {
		t.Fatalf("unexpected standardized head: %s", got.Head.String())
	}

	if got.Body[0].String() != "parent($7_X, $7_Y)" {
		t.Fatalf("unexpected first standardized goal: %s", got.Body[0].String())
	}

	if got.Body[1].String() != "parent($7_Y, $7_Z)" {
		t.Fatalf("unexpected second standardized goal: %s", got.Body[1].String())
	}

	if clause.Head.String() != "grandparent(X, Z)" {
		t.Fatalf("standardizeApart should not modify original clause: %s", clause.Head.String())
	}
}

func TestStandardizeApartCreatesDifferentVariablesForEachUse(t *testing.T) {
	clause := Rule(
		p("ancestor", Var("X"), Var("Y")),
		p("parent", Var("X"), Var("Y")),
	)

	first := standardizeApart(clause, 1)
	second := standardizeApart(clause, 2)

	if first.Head.String() == second.Head.String() {
		t.Fatalf(
			"expected separate clause uses to have distinct variables, got %s",
			first.Head.String(),
		)
	}
}
