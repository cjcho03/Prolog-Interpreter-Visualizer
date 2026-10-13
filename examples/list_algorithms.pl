% Local concatenation helper used by reverse_list/2.
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
    starts_with(NeedleTail, HaystackTail).
