# Prolog Interpreter

This project is a basic Prolog interpreter written in Go, with a small TypeScript/Vite visualizer for stepping through execution traces.

The goal is to build a small but understandable Prolog-style resolver from the ground up, focusing on the core execution model while gradually adding useful language features.

## Requirements

* Go 1.23.4 or a compatible newer version
* Node.js 20.19+ or 22.12+
* npm

The Go interpreter and CLI can be used without Node.js. Node.js is only required for the web visualizer and its end-to-end tests.

## Current Features

The interpreter currently supports:

* Atoms
* Numbers, including decimal values
* Variables
* Anonymous variables using `_`
* Compound terms
* Lists
* Predicates
* Facts and rules
* Clauses
* Unification
* Structural non-unification with `\=/2`
* Arithmetic evaluation with `is/2`
* Arithmetic expressions with `+`, `-`, `*`, `/`, `mod`, and `**`
* Arithmetic functions `abs/1`, `min/2`, and `max/2`
* Numeric comparison operators `<`, `>`, `=<`, `>=`, `=:=`, and `=\=`
* Arithmetic operator precedence and parentheses
* Multiple query goals
* Rule-body expansion
* Depth-first resolution
* Backtracking
* Recursive rules
* Variable standardization for separate rule calls
* Built-in predicates including `true/0`, `fail/0`, `=/2`, `\=/2`, `is/2`, and numeric comparisons
* Loading Prolog programs from `.pl` files
* A command-line interface with both one-shot queries and an interactive `?-` REPL
* A TypeScript/Vite visualizer for stepping through resolver execution
* Execution tracing for:
  * goals
  * built-in predicate execution
  * clause attempts
  * unification
  * rule expansion
  * failures
  * backtracking
  * solutions

The repository also includes unit, integration, CLI, and browser end-to-end tests covering:

* lexing and parsing
* unification
* facts and rules
* recursive resolution
* backtracking
* variable standardization
* anonymous variables
* numbers
* compound terms
* lists
* built-in predicates
* structural non-unification
* arithmetic evaluation
* numeric comparisons
* `.pl` file loading
* command-line execution
* interactive REPL queries
* example programs
* visualizer error handling and trace controls
* responsive browser behavior

## Parser Support

The project includes a parser for a small Prolog-like syntax.

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

Anonymous variables are supported when a value should be ignored:

```prolog
?- parent(alice, _).
?- parent(_, Who).
```

Number terms are supported as atomic values, including decimals:

```prolog
age(alice, 30).
temperature(room, 21.5).

?- age(alice, Age).
```

Compound terms are supported as nested values inside predicates:

```prolog
likes(alice, food(pizza)).
likes(bob, food(sushi)).

?- likes(alice, What).
?- likes(Person, food(pizza)).
```

List syntax includes empty lists, normal lists, and head-tail lists:

```prolog
items([]).
items([pizza, sushi, ramen]).

?- items([Head | Tail]).
?- items([pizza, sushi | Rest]).
```

Arithmetic expressions use normal operator precedence and may be evaluated with `is/2`:

```prolog
?- X is 2 + 3 * 4.
?- Y is (2 + 3) * 4.
?- Z is 5 / 2.
?- R is 10 mod 3.
?- P is 2 ** 4.
?- A is abs(-7).
?- M is max(4, 9).
```

Arithmetic expressions remain ordinary compound terms until explicitly evaluated. For example, `X = 2 + 3` binds `X` to the expression, while `X is 2 + 3` binds `X` to `5`. Exponentiation is right-associative, and `mod` requires integer operands.

Numeric comparisons evaluate both operands as arithmetic expressions:

```prolog
?- 3 < 5.
?- 5 >= 5.
?- 10 =:= 5 + 5.
?- 4 =\= 2 + 3.
```

`=:=` tests numeric equality after evaluation, while `=\=` tests numeric inequality. This is different from `=/2`, which performs structural unification without evaluating arithmetic.

Structural non-unification is available with `\=/2`. It succeeds only when two terms cannot be unified, and temporary bindings from the test are discarded:

```prolog
?- a \= b.
?- pair(a, b) \= pair(a, c).
?- X \= a.  % fails because X can unify with a
```

Parsed programs are converted into the Go clause representation and executed by the resolver.

## Example Programs

Several sample programs are included under `examples/`:

* `family.pl` - family facts and a two-goal rule
* `list_basics.pl` - `list_size/2`, `contains_item/2`, and `concat_lists/3`
* `list_algorithms.pl` - `reverse_list/2`, `suffix_from/3`, and `contiguous_slice/2`
* `sorted_merge.pl` - `merge_sorted/3`, a recursive sorted-list merge using numeric comparisons
* `arithmetic_eval.pl` - arithmetic evaluation through a user-defined `eval/2` rule

The examples are used by the CLI and integration tests.

## Command-Line Interface

The repository includes a command-line interpreter in `cmd/prolog`. It can load a `.pl` file, run a one-shot query, or start an interactive query loop while keeping the loaded program in memory.

### Interactive REPL

From the repository root:

```powershell
go run ./cmd/prolog examples/family.pl
```

Example session:

```text
Loaded 5 clauses from family.pl

?- parent(alice, X).
X = bob.
X = carol.

?- grandparent(alice, Who).
Who = diana.
Who = eli.

?- halt.
```

At the prompt, the leading `?-` is optional because the CLI adds it automatically when needed.

Use any of the following to exit:

```text
halt.
quit.
exit.
```

Parse errors are reported without terminating the session, so another query can be entered afterward.

### One-shot queries

A query can also be passed directly:

```powershell
go run ./cmd/prolog examples/family.pl "?- grandparent(alice, Who)."
```

Expected output:

```text
Loaded 5 clauses from family.pl

Who = diana.
Who = eli.
```

Ground queries print `true.` or `false.` depending on whether the query succeeds.

Queries should be quoted when they contain spaces or shell-sensitive operators such as `>` or `<`.

More examples:

```powershell
go run ./cmd/prolog examples/list_basics.pl "?- contains_item(Planet, [mercury,venus,earth])."
go run ./cmd/prolog examples/list_algorithms.pl "?- reverse_list([north,east,south], R)."
go run ./cmd/prolog examples/sorted_merge.pl "?- merge_sorted([2,6,10],[1,4,7,12], R)."
go run ./cmd/prolog examples/arithmetic_eval.pl "?- eval(14 + 3 * 5 - 8 / 2, Y)."
```

## Web Visualizer

The `visualizer/` directory contains a TypeScript/Vite interface for exploring resolver execution one event at a time.

The visualizer includes:

* selectable example programs
* editable program and query input
* solution output
* current goal and clause inspection
* variable bindings
* expanded rule goals
* clickable trace events
* previous/next stepping
* trace scrubbing
* automatic trace playback
* playback speed controls
* keyboard execution with `Ctrl+Enter` or `Cmd+Enter`
* responsive layout
* clear parser, query, and API error states

### Local development

Start the Go API from the repository root:

```powershell
go run ./cmd/server
```

Then start the visualizer in a second terminal:

```powershell
cd visualizer
npm install
npm run dev
```

The Vite development server proxies `/api` requests to:

```text
http://localhost:8080
```

For normal local development, `visualizer/.env` may contain:

```env
VITE_API_BASE_URL=
```

When the value is empty, the frontend uses same-origin `/api` requests, allowing the Vite proxy to forward them to the local Go server.

For a deployed frontend whose Go API is hosted separately, set the full API origin:

```env
VITE_API_BASE_URL=https://prolog-api.example.com
```

### Production build

From `visualizer/`:

```bash
npm install
npm run build
```

The production frontend is written to:

```text
visualizer/dist/
```

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

## List Example

```prolog
items([pizza, sushi, ramen]).
```

Query:

```prolog
?- items([Head | Tail]).
```

Expected answer:

```text
Head = pizza
Tail = [sushi, ramen]
```

## Arithmetic Example

The arithmetic example evaluates an expression through an ordinary user-defined rule:

```prolog
eval(E, V) :-
    V is E.
```

Example queries:

```prolog
?- eval(14 + 3 * 5 - 8 / 2, Y).
?- eval((14 + 3) * (5 - 2), Y).
?- eval(3 ** 2 ** 2, Y).
```

The integration tests also exercise parentheses, right-associative exponentiation, unary minus, `mod`, arithmetic functions, previously bound variables, and invalid arithmetic such as division by zero.

## Built-in Predicates

The resolver checks built-in predicates before searching user-defined clauses. Built-ins are dispatched by predicate name and arity, so `true/0` is distinct from a user-defined predicate such as `true/1`.

Examples:

```prolog
?- true.
?- fail.
?- X = alice.
?- a \= b.
?- X is 2 + 3 * 4.
?- 3 < 5.
?- 5 >= 5.
?- 10 =:= 5 + 5.
?- 4 =\= 2 + 3.
```

`true/0` succeeds immediately and continues with the remaining goals.

`fail/0` deliberately fails the current branch and causes normal backtracking.

`=/2` performs explicit structural unification without evaluating arithmetic expressions.

`\=/2` succeeds only when its arguments are not unifiable and does not create bindings.

`is/2` evaluates the arithmetic expression on its right and unifies the result with its left argument.

Numeric comparison built-ins evaluate both operands and succeed or fail without creating bindings.

Built-ins may also appear inside rule bodies:

```prolog
choice(first) :- fail.
choice(second).

?- choice(Value).
```

The query above returns `Value = second` after `fail/0` forces the first branch to backtrack.

## Testing

### Go unit and integration tests

From the repository root:

```bash
go test ./...
go vet ./...
```

The Go test suite covers the lexer, parser, unification, resolver, built-ins, arithmetic, comparisons, structural non-unification, recursive examples, `.pl` loading, and the command-line interface.

### Browser end-to-end tests

The visualizer includes a Playwright end-to-end test suite that exercises the real browser, Vite frontend, Go HTTP API, parser, resolver, and rendered execution trace.

From `visualizer/`:

```bash
npm install
npx playwright install chromium
npm run test:e2e
```

If Chromium requires additional Linux dependencies:

```bash
npx playwright install --with-deps chromium
```

The E2E suite covers:

* family queries and multiple solutions
* anonymous variables
* recursive list examples
* sorted merge and numeric comparisons
* arithmetic and built-ins
* structural and numeric inequality
* parser and query errors
* unavailable API handling
* trace stepping, scrubbing, jumping, and playback
* `Ctrl+Enter`
* reset behavior
* responsive layout

The current v1 suite contains 16 Playwright tests.

To open the Playwright HTML report after a test run:

```bash
npx playwright show-report
```

## Current Limitations

This project implements an educational subset of Prolog rather than full ISO Prolog.

The parser and resolver do not yet support:

* Strings
* Negation
* Cut (`!`)
* Disjunction
* If-then-else
* Meta-predicates such as `call/1`
* Term decomposition with `=../2`
* Many standard-library predicates
* Full Prolog syntax compatibility

The current implementation focuses on the core resolution model: terms, unification, facts and rules, depth-first search, backtracking, recursion, lists, arithmetic, comparisons, built-in predicates, and execution tracing.

## Planned Features

Future versions may add:

* Strings
* Negation
* Cut (`!`)
* Additional built-in and standard-library predicates
* Disjunction and other control-flow constructs
* Better parser error messages
* Improved answer formatting
* Broader Prolog syntax compatibility

## Project Direction

The project is now usable through three interfaces:

* the Go resolver API
* the command-line interpreter and interactive REPL
* the TypeScript execution visualizer

The current version focuses on making Prolog resolution understandable by exposing goals, clause attempts, unification, rule expansion, failures, backtracking, built-in execution, variable bindings, and solutions as trace events.

Future work can expand language compatibility without changing the core resolver and visualization model.
