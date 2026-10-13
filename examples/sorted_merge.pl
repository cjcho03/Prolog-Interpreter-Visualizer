% Merge two sorted numeric lists into one sorted list.
% This example exercises recursion, list unification, and numeric comparisons.

merge_sorted([], Right, Right).
merge_sorted(Left, [], Left).
merge_sorted([LeftHead|LeftTail], [RightHead|RightTail], [LeftHead|Result]) :-
    LeftHead =< RightHead,
    merge_sorted(LeftTail, [RightHead|RightTail], Result).
merge_sorted([LeftHead|LeftTail], [RightHead|RightTail], [RightHead|Result]) :-
    LeftHead > RightHead,
    merge_sorted([LeftHead|LeftTail], RightTail, Result).
