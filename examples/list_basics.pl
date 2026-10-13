% Small recursive list utilities used by the interpreter examples.

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
    concat_lists(Tail, Right, Result).
