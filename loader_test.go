package prolog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProgramFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "family.pl")
	program := `
parent(alice, bob).
parent(bob, charlie).
grandparent(X, Z) :- parent(X, Y), parent(Y, Z).	
`

	if err := os.WriteFile(path, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}

	clauses, err := LoadProgramFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(clauses) != 3 {
		t.Fatalf("expected 3 clauses, got %d", len(clauses))
	}

	if got := clauses[2].String(); got != "grandparent(X, Z) :- parent(X, Y), parent(Y, Z)" {
		t.Fatalf("unexpected parsed rule: %s", got)
	}
}

func TestLoadProgramFileReportsReadError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.pl")

	_, err := LoadProgramFile(path)
	if err == nil {
		t.Fatal("expected missing file to return an error")
	}

	if !strings.Contains(err.Error(), "read Prolog program") || !strings.Contains(err.Error(), "missing.pl") {
		t.Fatalf("expected contextual read error, got %q", err)
	}
}

func TestLoadProgramFileReportsParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.pl")

	if err := os.WriteFile(path, []byte("parent(alice, bob"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadProgramFile(path)
	if err == nil {
		t.Fatal("expected invalid Prolog source to return an error")
	}

	if !strings.Contains(err.Error(), "parse Prolog program") || !strings.Contains(err.Error(), "broken.pl") {
		t.Fatalf("expected contextual parse error, got %q", err)
	}
}

func TestLoadProgramFileRejectsEmptyPath(t *testing.T) {
	_, err := LoadProgramFile("")
	if err == nil {
		t.Fatal("expected empty path to return an error")
	}

	if !strings.Contains(err.Error(), "path cannot be empty") {
		t.Fatalf("unexpected empty-path error: %q", err)
	}
}
