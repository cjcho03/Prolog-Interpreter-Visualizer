export interface ExamplePreset {
  id: string
  name: string
  description: string
  program: string
  query: string
}

export const examplePresets: ExamplePreset[] = [
  {
    id: "family",
    name: "Family relationships",
    description: "Facts, a two-goal rule, unification, and multiple solutions.",
    program: `parent(alice, bob).
parent(alice, carol).
parent(bob, diana).
parent(carol, eli).

grandparent(X, Z) :-
    parent(X, Y),
    parent(Y, Z).`,
    query: "?- grandparent(alice, Who).",
  },
  {
    id: "list-basics",
    name: "List basics",
    description: "Recursive counting, membership, concatenation, arithmetic, and backtracking.",
    program: `% Small recursive list utilities used by the interpreter examples.

% list_size(List, Size)
% Counts top-level elements only.
list_size([], 0).
list_size([_|Tail], Size) :-
    list_size(Tail, Rest),
    Size is Rest + 1.

% contains_item(Item, List)
% Succeeds once for each matching position when Item is unbound.
contains_item(Item, [Item|_]).
contains_item(Item, [_|Tail]) :-
    contains_item(Item, Tail).

% concat_lists(Left, Right, Result)
concat_lists([], Right, Right).
concat_lists([Head|Tail], Right, [Head|Result]) :-
    concat_lists(Tail, Right, Result).`,
    query: "?- contains_item(Planet, [mercury, venus, earth]).",
  },
  {
    id: "list-algorithms",
    name: "List algorithms",
    description: "Reversal, suffix selection, contiguous slices, and deeper recursive traces.",
    program: `% Local concatenation helper used by reverse_list/2.
concat_lists([], Right, Right).
concat_lists([Head|Tail], Right, [Head|Result]) :-
    concat_lists(Tail, Right, Result).

% reverse_list(List, Reversed)
reverse_list([], []).
reverse_list([Head|Tail], Reversed) :-
    reverse_list(Tail, ReversedTail),
    concat_lists(ReversedTail, [Head], Reversed).

% suffix_from(List, Position, Suffix)
% Position is one-based; position 1 returns the original list.
suffix_from(List, 1, List).
suffix_from([_|Tail], Position, Suffix) :-
    Position > 1,
    Next is Position - 1,
    suffix_from(Tail, Next, Suffix).

% contiguous_slice(Slice, List)
% Succeeds when Slice appears contiguously inside List.
contiguous_slice([], _).
contiguous_slice([Head|SliceTail], [Head|ListTail]) :-
    starts_with(SliceTail, ListTail).
contiguous_slice(Slice, [_|ListTail]) :-
    contiguous_slice(Slice, ListTail).

starts_with([], _).
starts_with([Head|NeedleTail], [Head|HaystackTail]) :-
    starts_with(NeedleTail, HaystackTail).`,
    query: "?- reverse_list([north, east, south], Reversed).",
  },
  {
    id: "sorted-merge",
    name: "Sorted merge",
    description: "Recursive list processing with numeric comparisons and branch selection.",
    program: `% Merge two sorted numeric lists into one sorted list.
% This example exercises recursion, list unification, and numeric comparisons.

merge_sorted([], Right, Right).
merge_sorted(Left, [], Left).
merge_sorted([LeftHead|LeftTail], [RightHead|RightTail], [LeftHead|Result]) :-
    LeftHead =< RightHead,
    merge_sorted(LeftTail, [RightHead|RightTail], Result).
merge_sorted([LeftHead|LeftTail], [RightHead|RightTail], [RightHead|Result]) :-
    LeftHead > RightHead,
    merge_sorted([LeftHead|LeftTail], RightTail, Result).`,
    query: "?- merge_sorted([2, 6, 10], [1, 4, 7, 12], Result).",
  },
]
