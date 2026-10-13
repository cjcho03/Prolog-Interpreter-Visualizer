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
			name:  "mixed precedence",
			query: "?- eval(14 + 3 * 5 - 8 / 2, Y).",
			want:  Number("25.0"),
		},
		{
			name:  "parentheses override precedence",
			query: "?- eval((14 + 3) * (5 - 2), Y).",
			want:  Number("51"),
		},
		{
			name:  "right associative exponentiation",
			query: "?- eval(3 ** 2 ** 2, Y).",
			want:  Number("81"),
		},
		{
			name:  "unary minus with exponentiation",
			query: "?- eval(-3 ** 2, Y).",
			want:  Number("-9"),
		},
		{
			name:  "modulo",
			query: "?- eval(38 mod 7, Y).",
			want:  Number("3"),
		},
		{
			name:  "arithmetic functions",
			query: "?- eval(max(abs(-11), min(18, 7)), Y).",
			want:  Number("11"),
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
			query: "?- eval(27 / 0, Y).",
		},
		{
			name:  "unbound arithmetic variable",
			query: "?- eval(Input + 4, Y).",
		},
		{
			name:  "modulo by zero",
			query: "?- eval(27 mod 0, Y).",
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

	goals, err := ParseQuery("?- Base = 8, eval(Base * 4 - 3, Result).")
	if err != nil {
		t.Fatal(err)
	}

	answers := (Engine{Clauses: clauses}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected exactly 1 answer, got %d", len(answers))
	}

	base := resolveTerm(Var("Base"), answers[0])
	result := resolveTerm(Var("Result"), answers[0])

	if base != Number("8") {
		t.Fatalf("expected Base = 8, got %v", base)
	}

	if result != Number("29") {
		t.Fatalf("expected Result = 29, got %v", result)
	}
}
