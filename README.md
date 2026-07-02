# Prolog Interpreter

This project is a basic Prolog interpreter written in Go.

Right now, it only includes the core pieces needed to get a simple interpreter working:

* Atoms
* Variables
* Predicates
* Facts
* Unification
* Multiple query goals
* Depth-first search
* Basic backtracking
* Tests for unification and query resolution

The current version does not yet parse normal Prolog source code. Facts and queries are created directly in Go so the underlying logic can be tested first.

## Planned Features

Future versions may add:

* Rules, such as `grandparent(X, Z) :- parent(X, Y), parent(Y, Z).`
* A parser for `.pl` Prolog files
* A command-line interface
* Interactive queries using `?-`
* Built-in predicates
* Lists
* Anonymous variables using `_`
* Negation
* Cut (`!`)
* Recursion support through rules
* Better error handling
* Query tracing to show unification and backtracking steps
* More complete test coverage
