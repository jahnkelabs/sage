package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpIncludesAliasesSectionFromDump(t *testing.T) {
	resetAliasLoaderForTest()

	abs, err := filepath.Abs(filepath.Join("..", "testdata", "merged-compose-example.yml"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SAGE_COMPOSE_CONFIG_DUMP", abs)

	buf := new(bytes.Buffer)
	root := newRootCmd("v", "abc", "d")
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "migrate") || !strings.Contains(out, "Aliases") {
		t.Fatalf("help output missing aliases:\n%s", out)
	}
}
