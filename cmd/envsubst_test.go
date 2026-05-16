package cmd

import (
	"strings"
	"testing"
)

func TestExpandEnv_missingWarns(t *testing.T) {
	t.Setenv("SAGE_EXPAND_TEST_X", "ok")
	var warns []string
	out := expandEnv("hello $SAGE_EXPAND_TEST_X ${MISSING_ONE}", func(s string) { warns = append(warns, s) })
	if out != "hello ok " {
		t.Fatalf("got %q", out)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "MISSING_ONE") {
		t.Fatalf("warns=%v", warns)
	}
}
