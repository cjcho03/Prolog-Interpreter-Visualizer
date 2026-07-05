package main

import (
	"log"
	"net/http"

	prolog "prolog-interpreter"
)

func main() {
	engine := prolog.Engine{
		Facts: []prolog.Predicate{
			{
				Name: "parent",
				Args: []prolog.Term{
					prolog.Atom("alice"),
					prolog.Atom("bob"),
				},
			},
			{
				Name: "parent",
				Args: []prolog.Term{
					prolog.Atom("alice"),
					prolog.Atom("carol"),
				},
			},
			{
				Name: "parent",
				Args: []prolog.Term{
					prolog.Atom("carol"),
					prolog.Atom("eli"),
				},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/demo", prolog.DemoHandler(engine))

	log.Println("Go API running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
