package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAliases_multiPerService(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "testdata", "merged-compose-example.yml"))
	if err != nil {
		t.Fatal(err)
	}
	idx, err := parseAliasesFromYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx) != 3 {
		t.Fatalf("expected 3 aliases, got %d", len(idx))
	}
	m := idx["migrate"]
	if m.Service != "api" || m.RawCmd != "bundle exec rake db:migrate" {
		t.Fatalf("migrate alias: %+v", m)
	}
	if len(m.Profiles) != 1 || m.Profiles[0] != "dev" {
		t.Fatalf("expected api profiles [dev], got %#v", m.Profiles)
	}
}

func TestParseAliases_duplicateRejected(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "testdata", "duplicate-alias.yml"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = parseAliasesFromYAML(raw)
	if err == nil {
		t.Fatal("expected duplicate alias error")
	}
}
