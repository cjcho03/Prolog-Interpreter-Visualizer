// unify_test.go
package prolog

import "testing"

func TestUnifyVariableWithAtom(t *testing.T) {
	sub := Substitution{}
	ok := unify(Var("Who"), Atom("bob"), sub)

	if !ok {
		t.Fatal("expected unification to succeed")
	}
	if sub[Var("Who")] != Atom("bob") {
		t.Fatalf("expected Who = bob, got %v", sub[Var("Who")])
	}
}

func TestUnifyMatchingAtoms(t *testing.T) {
	sub := Substitution{}
	ok := unify(Atom("alice"), Atom("alice"), sub)
	if !ok {
		t.Fatal("expected matching atoms to unify")
	}
	if len(sub) != 0 {
		t.Fatalf("expected no bindings, got %v", sub)
	}
}

func TestUnifyDifferentAtomsFails(t *testing.T) {
	sub := Substitution{}
	ok := unify(Atom("alice"), Atom("bob"), sub)

	if ok {
		t.Fatal("expected different atoms not to unify")
	}
}

func TestUnifySameVariable(t *testing.T) {
	sub := Substitution{}
	ok := unify(Var("X"), Var("X"), sub)

	if !ok {
		t.Fatal("expected X = X to succeed")
	}

	if len(sub) != 0 {
		t.Fatalf("expected no self-binding, got %v", sub)
	}
}

func TestDereferenceVariableChain(t *testing.T) {
	sub := Substitution{}
	goal := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Var("Who")},
	}

	fact := Predicate{
		Name: "parent",
		Args: []Term{Atom("alice"), Atom("bob")},
	}

	ok := unifyPredicate(goal, fact, sub)
	if !ok {
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
