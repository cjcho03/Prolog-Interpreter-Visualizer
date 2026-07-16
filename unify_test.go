// unify_test.go
package prolog

import "testing"

func TestUnifyVariableWithAtom(t *testing.T) {
	sub := Substitution{}

	if !unify(Var("Who"), Atom("bob"), sub) {
		t.Fatal("expected unification to succeed")
	}

	if sub[Var("Who")] != Atom("bob") {
		t.Fatalf("expected Who = bob, got %v", sub[Var("who")])
	}
}

func TestUnifyAtomWithVariable(t *testing.T) {
	sub := Substitution{}

	if !unify(Var("Who"), Atom("bob"), sub) {
		t.Fatal("expected unification to succeed")
	}
	if sub[Var("Who")] != Atom("bob") {
		t.Fatalf("expected Who = bob, got %v", sub[Var("Whos")])
	}
}

func TestUnifyMatchingAtoms(t *testing.T) {
	sub := Substitution{}

	if !unify(Atom("alice"), Atom("alice"), sub) {
		t.Fatal("expected matching atoms to unify")
	}

	if len(sub) != 0 {
		t.Fatalf("expected no bindings, got %v", sub)
	}
}

func TestUnifyDifferentAtomsFails(t *testing.T) {
	sub := Substitution{}

	if unify(Atom("alice"), Atom("bob"), sub) {
		t.Fatalf("expectd no bindings, got %v", sub)
	}
}

func TestUnifyRespectsExistingBinding(t *testing.T) {
	sub := Substitution{
		Var("X"): Atom("alice"),
	}

	if unify(Var("X"), Atom("bob"), sub) {
		t.Fatal("expected X = bob to fail when X is already alice")
	}
}

func TestUnifySameVariable(t *testing.T) {
	sub := Substitution{}

	if !unify(Var("X"), Var("X"), sub) {
		t.Fatal("expected X = X to succeed")
	}

	if len(sub) != 0 {
		t.Fatalf("expected no self-binding, got %v", sub)
	}
}

func TestDereferenceVariableChain(t *testing.T) {
	sub := Substitution{
		Var("X"): Var("Y"),
		Var("Y"): Atom("alice"),
	}

	got := dereference(Var("X"), sub)

	if got != Atom("alice") {
		t.Fatalf("expected X to resolve to alice, got %v", got)
	}
}

func TestUnifyPredicate(t *testing.T) {
	sub := Substitution{}

	goal := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Var("Who")},
	}

	fact := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Atom("bob")},
	}

	if !unifyPredicate(goal, fact, sub) {
		t.Fatal("expected predicates to unify")
	}

	if sub[Var("Who")] != Atom("bob") {
		t.Fatalf("expected Who = bob, got %v", sub[Var("Who")])
	}
}

func TestUnifyPredicateDifferentNameFails(t *testing.T) {
	sub := Substitution{}

	goal := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Var("Who")},
	}

	fact := Predicate{
		Name: "likes",
		Args: []Term{Atom("alice"), Atom("bob")},
	}

	if unifyPredicate(goal, fact, sub) {
		t.Fatal("expected predicates with different names not to unify")
	}
}

func TestUnifyPredicateDifferentArityFails(t *testing.T) {
	sub := Substitution{}

	goal := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Var("Who")},
	}

	fact := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice")},
	}

	if unifyPredicate(goal, fact, sub) {
		t.Fatal("expected predicates with different arity not to unify")
	}
}

func TestUnifyMatchingNumbers(t *testing.T) {
	sub := Substitution{}

	if !unify(Number("30"), Number("30"), sub) {
		t.Fatal("expected matching numbers to unify")
	}

	if len(sub) != 0 {
		t.Fatalf("expected no bindings, got %v", sub)
	}
}

func TestUnifyDifferentNumbersFails(t *testing.T) {
	sub := Substitution{}

	if unify(Number("30"), Number("24"), sub) {
		t.Fatalf("expected different numbers not to unify, got %v", sub)
	}
}

func TestUnifyVariableWithNumber(t *testing.T) {
	sub := Substitution{}

	if !unify(Var("Age"), Number("30"), sub) {
		t.Fatal("expected variable to unify with number")
	}

	if sub[Var("Age")] != Number("30") {
		t.Fatalf("expected Age = 30, got %v", sub[Var("Age")])
	}
}
