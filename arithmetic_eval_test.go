package prolog

import "testing"

func TestArithmeticEvalExample(t *testing.T) {
	clauses, err := LoadProgramFile("examples/arithmetic_eval.pl")
	if err != nil {
		t.Fatalf("load arithmetic example: %v", err)
	}

	tests := []struct {
		name  string
		query string
		want  Number
	}{
		{
			name:  "mixed precedence with negative result",
			query: "?- eval(5 - 6 * 18 / 3 + 2, Y).",
			want:  Number("-29.0"),
		},
		{
			name:  "mixed multiplication division addition",
			query: "?- eval(10 * 20 - 9 / 3 + 20, Y).",
			want:  Number("217.0"),
		},
		{
			name:  "power multiplication subtraction",
			query: "?- eval(10 ** 3 * 9 - 100, Y).",
			want:  Number("8900"),
		},
		{
			name:  "parentheses override precedence",
			query: "?- eval((5 - 6) * (18 / 3 + 2), Y).",
			want:  Number("-8.0"),
		},
		{
			name:  "right associative exponentiation",
			query: "?- eval(2 ** 3 ** 2, Y).",
			want:  Number("512"),
		},
		{
			name:  "unary minus with exponentiation",
			query: "?- eval(-2 ** 2, Y).",
			want:  Number("-4"),
		},
		{
			name:  "modulo",
			query: "?- eval(29 mod 6, Y).",
			want:  Number("5"),
		},
		{
			name:  "arithmetic functions",
			query: "?- eval(max(abs(-7), min(12,9)), Y).",
			want:  Number("9"),
		},
	}

	engine := Engine{Clauses: clauses}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goals, err := ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("parse query %q: %v", tc.query, err)
			}

			answers := engine.Solve(goals...)
			if len(answers) != 1 {
				t.Fatalf("expected exactly 1 answer for %q, got %d", tc.query, len(answers))
			}

			got := resolveTerm(Var("Y"), answers[0])

			number, ok := got.(Number)
			if !ok {
				t.Fatalf("expected Y to resolve to Number, got %T (%v)", got, got)
			}

			if number != tc.want {
				t.Fatalf("for %q expected Y = %s, got %s", tc.query, tc.want, number)
			}
		})
	}
}

func TestArithmeticEvalExampleFailsForInvalidArithmetic(t *testing.T) {
	clauses, err := LoadProgramFile("examples/arithmetic_eval.pl")
	if err != nil {
		t.Fatalf("load arithmetic example: %v", err)
	}

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "division by zero",
			query: "?- eval(10 / 0, Y).",
		},
		{
			name:  "unbound arithmetic variable",
			query: "?- eval(X + 1, Y).",
		},
		{
			name:  "modulo by zero",
			query: "?- eval(10 mod 0, Y).",
		},
	}

	engine := Engine{Clauses: clauses}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goals, err := ParseQuery(tc.query)
			if err != nil {
				t.Fatalf("parse query %q: %v", tc.query, err)
			}

			answers := engine.Solve(goals...)
			if len(answers) != 0 {
				t.Fatalf("expected %q to fail, got %d answers", tc.query, len(answers))
			}
		})
	}
}

func TestArithmeticEvalExampleCanUsePreviouslyBoundVariable(t *testing.T) {
	clauses, err := LoadProgramFile("examples/arithmetic_eval.pl")
	if err != nil {
		t.Fatalf("load arithmetic example: %v", err)
	}

	goals, err := ParseQuery("?- X = 7, eval(X * 3 + 1, Y).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{Clauses: clauses}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected exactly 1 answer, got %d", len(answers))
	}

	x := resolveTerm(Var("X"), answers[0])
	y := resolveTerm(Var("Y"), answers[0])

	if x != Number("7") {
		t.Fatalf("expected X = 7, got %v", x)
	}

	if y != Number("22") {
		t.Fatalf("expected Y = 22, got %v", y)
	}
}
