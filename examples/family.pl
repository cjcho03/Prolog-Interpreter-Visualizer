parent(alice, bob).
parent(alice, carol).
parent(bob, diana).
parent(carol, eli).

grandparent(X, Z) :-
    parent(X, Y),
    parent(Y, Z).
