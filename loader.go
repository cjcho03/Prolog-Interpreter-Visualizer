package prolog

import (
	"fmt"
	"os"
)

// LoadProgramFile reads a Prolog source file from disk and parses its clauses.
// The file is expected to contain program clauses rather than a ?- query.
func LoadProgramFile(path string) ([]Clause, error) {
	if path == "" {
		return nil, fmt.Errorf("program path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Prolog program %q: %w", path, err)
	}

	clauses, err := ParseProgram(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse Prolog program %q: %w", path, err)
	}

	return clauses, nil
}
