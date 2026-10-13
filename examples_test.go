package prolog

import (
	"path/filepath"
	"testing"
)

func loadExampleEngine(t *testing.T, name string) Engine {
	t.Helper()

	clauses, err := LoadProgramFile(filepath.Join("examples", name))
	if err != nil {
		t.Fatalf("load example %q: %v", name, err)
	}

	return Engine{Clauses: clauses}
}

func solveExampleQuery(t *testing.T, engine Engine, query string) []Substitution {
	t.Helper()

	goals, err := ParseQuery(query)
	if err != nil {
		t.Fatalf("parse query %q: %v", query, err)
	}

	return engine.Solve(goals...)
}

func requireExampleBinding(t *testing.T, answer Substitution, variable Var, want string) {
	t.Helper()

	got := resolveTerm(variable, answer).String()
	if got != want {
		t.Fatalf("expected %s = %s, got %s", variable, want, got)
	}
}

func TestListBasicsExample(t *testing.T) {
	engine := loadExampleEngine(t, "list_basics.pl")

	t.Run("counts top level values", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- list_size([oak,[pine,maple],cedar], N).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("N"), "3")
	})

	t.Run("counts a longer flat list", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- list_size([north,east,south,west], N).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("N"), "4")
	})

	t.Run("ground membership succeeds", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- contains_item(venus, [mercury,venus,earth]).")
		if len(answers) == 0 {
			t.Fatal("expected contains_item/2 to succeed")
		}
	})

	t.Run("membership enumerates values", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- contains_item(Planet, [mercury,venus,earth]).")
		if len(answers) != 3 {
			t.Fatalf("expected 3 answers, got %d", len(answers))
		}

		wants := []string{"mercury", "venus", "earth"}
		for i, want := range wants {
			requireExampleBinding(t, answers[i], Var("Planet"), want)
		}
	})

	t.Run("concatenates two lists", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- concat_lists([red,green],[blue,yellow], Combined).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Combined"), "[red, green, blue, yellow]")
	})

	t.Run("concatenates with an empty right side", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- concat_lists([spring,summer],[], Combined).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Combined"), "[spring, summer]")
	})
}

func TestListAlgorithmsExample(t *testing.T) {
	engine := loadExampleEngine(t, "list_algorithms.pl")

	t.Run("reverses a list", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- reverse_list([north,east,south], Reversed).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Reversed"), "[south, east, north]")
	})

	t.Run("returns the first suffix", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- suffix_from([p,q,r,s,t],1, Suffix).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Suffix"), "[p, q, r, s, t]")
	})

	t.Run("returns a later suffix", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- suffix_from([p,q,r,s,t],2, Suffix).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Suffix"), "[q, r, s, t]")
	})

	t.Run("finds a contiguous slice", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- contiguous_slice([gamma,delta],[alpha,beta,gamma,delta,epsilon]).")
		if len(answers) == 0 {
			t.Fatal("expected contiguous_slice/2 to succeed")
		}
	})

	t.Run("rejects a noncontiguous selection", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- contiguous_slice([alpha,gamma],[alpha,beta,gamma,delta]).")
		if len(answers) != 0 {
			t.Fatalf("expected no answers, got %d", len(answers))
		}
	})

	t.Run("binds a variable inside a slice", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- contiguous_slice([gamma,delta],[alpha,beta,gamma,X,epsilon]).")
		if len(answers) == 0 {
			t.Fatal("expected at least 1 answer")
		}
		requireExampleBinding(t, answers[0], Var("X"), "delta")
	})
}

func TestSortedMergeExample(t *testing.T) {
	engine := loadExampleEngine(t, "sorted_merge.pl")

	t.Run("merges interleaved numeric lists", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- merge_sorted([2,6,10],[1,4,7,12], Merged).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Merged"), "[1, 2, 4, 6, 7, 10, 12]")
	})

	t.Run("merges with an empty right list", func(t *testing.T) {
		answers := solveExampleQuery(t, engine, "?- merge_sorted([3,5,9],[], Merged).")
		if len(answers) != 1 {
			t.Fatalf("expected 1 answer, got %d", len(answers))
		}
		requireExampleBinding(t, answers[0], Var("Merged"), "[3, 5, 9]")
	})
}
