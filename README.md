# Prolog Interpreter

This project is a basic Prolog interpreter written in Go, with a small TypeScript visualizer for stepping through execution traces.

The goal is to build a small but understandable Prolog-style resolver from the ground up, starting with the core execution model before gradually adding more language features.

## Current Features

The interpreter currently supports:

* Atoms
* Numbers
* Variables
* Anonymous variables using `_`
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
* A visualizer that accepts user-entered Prolog programs and queries
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

Anonymous variables are also supported when a value should be ignored:

```prolog
?- parent(alice, _).
?- parent(_, Who).
```

Number terms are supported as atomic values:

```prolog
age(alice, 30).
age(bob, 24).

?- age(alice, Age).
```

Parsed programs can be converted into the existing Go clause representation and executed by the resolver.

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
* Strings
* Nested compound terms such as `food(pizza)`
* Arithmetic
* Built-in predicates
* Negation
* Cut (`!`)
* Operator precedence
* Full Prolog syntax compatibility

For now, terms are limited to atoms, numbers, and variables, matching the current unification model.

## Planned Features

Future versions may add:

* A command-line interface
* Loading `.pl` files from disk
* Interactive queries using `?-`
* Built-in predicates
* Lists
* Nested compound terms
* Strings
* Negation
* Cut (`!`)
* Better parser error messages
* Better answer formatting
* More complete test coverage

## Project Direction

The project has moved from a Go-only internal representation to a usable Prolog-like interpreter with parser support, a solve API, and a visualizer for user-entered programs and queries.

The next major milestones are expanding the supported term model, improving parser errors, and adding more Prolog language features such as nested compound terms, lists, strings, and built-in predicates.