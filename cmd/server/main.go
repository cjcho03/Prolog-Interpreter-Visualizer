package main

import (
	"log"
	"net/http"

	prolog "prolog-interpreter"
)

func predicate(name string, args ...prolog.Term) prolog.Predicate {
	return prolog.Predicate{
		Name: name,
		Args: args,
	}
}

func main() {
	engine := prolog.Engine{
		Clauses: []prolog.Clause{
			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("alice"),
					prolog.Atom("bob"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("alice"),
					prolog.Atom("carol"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("bob"),
					prolog.Atom("diana"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("carol"),
					prolog.Atom("eli"),
				),
			),

			prolog.Rule(
				predicate(
					"grandparent",
					prolog.Var("X"),
					prolog.Var("Z"),
				),
				predicate(
					"parent",
					prolog.Var("X"),
					prolog.Var("Y"),
				),
				predicate(
					"parent",
					prolog.Var("Y"),
					prolog.Var("Z"),
				),
			),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/demo", prolog.DemoHandler(engine))
	mux.HandleFunc("/api/solve", prolog.SolveHandler())

	log.Println("Go API running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
