package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	prolog "prolog-interpreter"
)

const usage = `Usage: prolog <program.pl> [query]

Load a Prolog program from disk.

With a query argument, execute one query and exit.
Without a query argument, start an interactive ?- prompt.

Examples:
  prolog family.pl
  prolog family.pl "?- grandparent(alice, Who)."

Interactive commands:
  halt.   exit the REPL
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	if args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, usage)
		return 0
	}

	programPath := args[0]
	clauses, err := prolog.LoadProgramFile(programPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(
		stdout,
		"Loaded %d clauses from %s\n",
		len(clauses),
		filepath.Base(programPath),
	)

	engine := prolog.Engine{Clauses: clauses}

	if len(args) == 1 {
		fmt.Fprintln(stdout)
		return runREPL(engine, stdin, stdout, stderr)
	}

	queryText := strings.TrimSpace(strings.Join(args[1:], " "))
	goals, err := prolog.ParseQuery(queryText)
	if err != nil {
		fmt.Fprintf(stderr, "error: parse query: %v\n", err)
		return 1
	}

	answers := engine.Solve(goals...)

	fmt.Fprintln(stdout)
	writeAnswers(stdout, goals, answers)
	return 0
}

func runREPL(
	engine prolog.Engine,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	scanner := bufio.NewScanner(stdin)

	for {
		fmt.Fprint(stdout, "?- ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(stderr, "error: read query: %v\n", err)
				return 1
			}

			return 0
		}

		queryText := strings.TrimSpace(scanner.Text())
		if queryText == "" {
			continue
		}

		if queryText == "halt." || queryText == "quit." || queryText == "exit." {
			return 0
		}

		if !strings.HasPrefix(queryText, "?-") {
			queryText = "?- " + queryText
		}

		goals, err := prolog.ParseQuery(queryText)
		if err != nil {
			fmt.Fprintf(stderr, "error: parse query: %v\n", err)
			continue
		}

		answers := engine.Solve(goals...)
		writeAnswers(stdout, goals, answers)
	}
}

func writeAnswers(
	writer io.Writer,
	goals []prolog.Predicate,
	answers []prolog.Substitution,
) {
	if len(answers) == 0 {
		fmt.Fprintln(writer, "false.")
		return
	}

	variables := queryVariables(goals)
	if len(variables) == 0 {
		fmt.Fprintln(writer, "true.")
		return
	}

	for _, answer := range answers {
		bindings := make([]string, 0, len(variables))

		for _, variable := range variables {
			resolved := resolveAnswerTerm(variable, answer, make(map[prolog.Var]bool))

			if resolvedVariable, ok := resolved.(prolog.Var); ok && resolvedVariable == variable {
				continue
			}

			bindings = append(
				bindings,
				fmt.Sprintf("%s = %s", variable.String(), resolved.String()),
			)
		}

		if len(bindings) == 0 {
			fmt.Fprintln(writer, "true.")
			continue
		}

		fmt.Fprintf(writer, "%s.\n", strings.Join(bindings, ", "))
	}
}

func queryVariables(goals []prolog.Predicate) []prolog.Var {
	seen := make(map[prolog.Var]bool)
	variables := make([]prolog.Var, 0)

	var collect func(prolog.Term)
	collect = func(term prolog.Term) {
		switch value := term.(type) {
		case prolog.Var:
			// ParseQuery gives anonymous variables internal names beginning with '$.'
			if strings.HasPrefix(string(value), "$") || seen[value] {
				return
			}

			seen[value] = true
			variables = append(variables, value)

		case prolog.Compound:
			for _, arg := range value.Args {
				collect(arg)
			}
		}
	}

	for _, goal := range goals {
		for _, arg := range goal.Args {
			collect(arg)
		}
	}

	return variables
}

func resolveAnswerTerm(
	term prolog.Term,
	sub prolog.Substitution,
	visiting map[prolog.Var]bool,
) prolog.Term {
	switch value := term.(type) {
	case prolog.Var:
		if visiting[value] {
			return value
		}

		next, found := sub[value]
		if !found {
			return value
		}

		visiting[value] = true
		resolved := resolveAnswerTerm(next, sub, visiting)
		delete(visiting, value)
		return resolved

	case prolog.Compound:
		args := make([]prolog.Term, len(value.Args))
		for i, arg := range value.Args {
			args[i] = resolveAnswerTerm(arg, sub, visiting)
		}

		return prolog.Compound{
			Name: value.Name,
			Args: args,
		}

	default:
		return term
	}
}
