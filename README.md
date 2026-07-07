# Prolog Interpreter

This project is a basic Prolog interpreter written in Go, with a small TypeScript visualizer for stepping through execution traces.

Right now, it includes the core pieces needed for a working Prolog-style resolver:

* Atoms
* Variables
* Predicates
* Facts and rules
* Clauses
* Unification
* Multiple query goals
* Rule-body expansion
* Depth-first resolution
* Backtracking
* Recursive rules
* Variable standardization for separate rule calls
* Execution tracing for goals, clause attempts, unification, rule expansion, failures, backtracking, and solutions
* A fixed visualizer demo for showing the resolution process
* Tests for unification, facts, rules, backtracking, and recursive resolution

The current version does not yet parse normal Prolog source code. Facts, rules, and queries are created directly in Go so the interpreter logic can be developed and tested before adding parsing.

## Planned Features

Future versions may add:

* A parser for `.pl` Prolog files
* A command-line interface
* Interactive queries using `?-`
* User-entered programs and queries through the visualizer
* Built-in predicates
* Lists
* Anonymous variables using `_`
* Negation
* Cut (`!`)
* Better error handling
* More complete test coverage