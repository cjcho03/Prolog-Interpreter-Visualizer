package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	prolog "prolog-interpreter"
)

func writeTestProgram(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "family.pl")
	program := `
parent(alice, bob).
parent(alice, carol).
parent(bob, diana).
parent(carol, eli).
grandparent(X, Z) :- parent(X, Y), parent(Y, Z).
`

	if err := os.WriteFile(path, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestRunLoadsProgramWithoutQuery(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{programPath}, strings.NewReader(""), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	if got := stdout.String(); got != "Loaded 5 clauses from family.pl\n\n?- " {
		t.Fatalf("unexpected output: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunExecutesQueryFromLoadedProgram(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{programPath, "?- grandparent(alice, Who)."},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	want := "Loaded 5 clauses from family.pl\n\nWho = diana.\nWho = eli.\n"
	if got := stdout.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRunPrintsFalseForQueryWithoutSolutions(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{programPath, "?- parent(nobody, Who)."},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	if !strings.HasSuffix(stdout.String(), "\nfalse.\n") {
		t.Fatalf("expected false result, got %q", stdout.String())
	}
}

func TestRunPrintsTrueForGroundQuery(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{programPath, "?- parent(alice, bob)."},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	if !strings.HasSuffix(stdout.String(), "\ntrue.\n") {
		t.Fatalf("expected true result, got %q", stdout.String())
	}
}

func TestRunJoinsSplitQueryArguments(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{programPath, "?-", "grandparent(alice,", "Who)."},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Who = diana.") ||
		!strings.Contains(stdout.String(), "Who = eli.") {
		t.Fatalf("expected both grandparent answers, got %q", stdout.String())
	}
}

func TestRunReportsProgramLoadError(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"missing.pl"}, strings.NewReader(""), &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "read Prolog program") {
		t.Fatalf("unexpected error output: %q", stderr.String())
	}
}

func TestRunReportsQueryParseError(t *testing.T) {
	programPath := writeTestProgram(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{programPath, "parent(alice, Who)."},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "parse query") {
		t.Fatalf("unexpected error output: %q", stderr.String())
	}
}

func TestRunWithoutArgumentsPrintsUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, strings.NewReader(""), &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "Usage: prolog <program.pl> [query]") {
		t.Fatalf("expected usage text, got %q", stderr.String())
	}
}

func TestRunHelpPrintsUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, strings.NewReader(""), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout.String(), "Usage: prolog <program.pl> [query]") {
		t.Fatalf("expected usage text, got %q", stdout.String())
	}
}

func TestRunInteractiveREPLExecutesMultipleQueries(t *testing.T) {
	programPath := writeTestProgram(t)

	input := strings.NewReader(
		"parent(alice, X).\n" +
			"?- grandparent(alice, Who).\n" +
			"parent(nobody, X).\n" +
			"halt.\n",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{programPath}, input, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d: %s", exitCode, stderr.String())
	}

	got := stdout.String()
	for _, want := range []string{
		"X = bob.",
		"X = carol.",
		"Who = diana.",
		"Who = eli.",
		"false.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected REPL output to contain %q, got %q", want, got)
		}
	}

	if strings.Count(got, "?- ") != 4 {
		t.Fatalf("expected one prompt per interactive command, got %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunInteractiveREPLContinuesAfterParseError(t *testing.T) {
	programPath := writeTestProgram(t)
	input := strings.NewReader(
		"parent(alice, X)\n" +
			"parent(alice, X).\n" +
			"halt.\n",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{programPath}, input, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "parse query") {
		t.Fatalf("expected parse error, got %q", stderr.String())
	}

	if !strings.Contains(stdout.String(), "X = bob.") ||
		!strings.Contains(stdout.String(), "X = carol.") {
		t.Fatalf("expected REPL to continue after parse error, got %q", stdout.String())
	}
}

func TestWriteAnswersResolvesVariableChains(t *testing.T) {
	goals, err := prolog.ParseQuery("?- X = Y, Y = alice.")
	if err != nil {
		t.Fatal(err)
	}

	answers := (prolog.Engine{}).Solve(goals...)
	if len(answers) != 1 {
		t.Fatalf("expected one answer, got %d", len(answers))
	}

	var output bytes.Buffer
	writeAnswers(&output, goals, answers)

	if got := output.String(); got != "X = alice, Y = alice.\n" {
		t.Fatalf("unexpected resolved bindings: %q", got)
	}
}
