// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// fixturesDir holds optional JSON overrides for the built-in responses. It is
// resolved relative to the working directory; the Makefile targets `cd` into
// this package before running, and record mode writes here.
const fixturesDir = "testdata/fixtures"

// loadOverride reads fixturesDir/<name>.json into v when the file exists.
// Returns (false, nil) if there is no override. A file that does not fit v's
// type is a hard error — that is the fixture drift guard, also asserted in
// fixtures_test.go.
func loadOverride(name string, v any) (bool, error) {
	path := filepath.Join(fixturesDir, name+".json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return false, fmt.Errorf("fixture %s does not fit its model type (regenerate it with `make keppel-mock-record`): %w", path, err)
	}
	return true, nil
}

// mustOverride applies an override fixture if present, logging and ignoring a
// missing file but failing fast on a malformed one.
func mustOverride(name string, v any) {
	used, err := loadOverride(name, v)
	if err != nil {
		log.Fatalf("keppel mock: %v", err)
	}
	if used {
		log.Printf("keppel mock: serving %s from %s/%s.json", name, fixturesDir, name)
	}
}

// writeFixture persists v as pretty JSON under fixturesDir (used by record mode).
func writeFixture(name string, b []byte) error {
	if err := os.MkdirAll(fixturesDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(fixturesDir, name+".json"), b, 0o644)
}
