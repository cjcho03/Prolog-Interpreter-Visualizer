# Prolog Interpreter

This project is a basic Prolog interpreter written in Go, with a small TypeScript visualizer for stepping through execution traces.

The goal is to build a small but understandable Prolog-style resolver from the ground up, starting with the core execution model before gradually adding more language features.

## Current Features

The interpreter currently supports:

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
* Execution tracing for:

  * goals
  * clause attempts
  * unification
  * rule expansion
  * failures
  * backtracking
  * solutions
* A fixed visualizer demo for showing the resolution process
* Tests for:

  * unification
  * facts
  * rules
  * backtracking
  * recursive resolution
  * variable standardization

## Parser Support

The project now includes an initial parser for a small Prolog-like syntax.

Supported program syntax includes facts:

```prolog
parent(alice, bob).
parent(bob, charlie).
```

and rules:

```prolog
grandparent(X, Z) :-
    parent(X, Y),
    parent(Y, Z).
```

Supported query syntax includes single-goal and multi-goal queries:

```prolog
?- parent(alice, Who).
?- parent(alice, X), parent(X, Y).
```

Parsed programs can be converted into the existing Go clause representation and executed by the resolver.

## Example

```prolog
parent(alice, bob).
parent(alice, carol).
parent(bob, diana).
parent(carol, eli).

grandparent(X, Z) :-
    parent(X, Y),
    parent(Y, Z).
```

Query:

```prolog
?- grandparent(alice, Who).
```

Expected answers:

```text
Who = diana
Who = eli
```

## Current Limitations

This is still a small educational interpreter, not a full Prolog implementation.

The current parser does not yet support:

* Lists
* Numbers
* Strings
* Anonymous variables using `_`
* Nested compound terms such as `food(pizza)`
* Arithmetic
* Built-in predicates
* Negation
* Cut (`!`)
* Operator precedence
* Full Prolog syntax compatibility

For now, terms are limited to atoms and variables, matching the current unification model.

## Planned Features

Future versions may add:

* A command-line interface
* Loading `.pl` files from disk
* Interactive queries using `?-`
* User-entered programs and queries through the visualizer
* Built-in predicates
* Lists
* Anonymous variables using `_`
* Nested compound terms
* Numbers and strings
* Negation
* Cut (`!`)
* Better parser error messages
* Better answer formatting
* More complete test coverage

## Project Direction

The project is currently moving from a Go-only internal representation toward a usable Prolog-like interpreter.

The next major milestone is replacing the fixed visualizer demo with a real solve endpoint that accepts user-provided Prolog source code and a query, parses both, runs the resolver, and returns answers plus execution trace events.
