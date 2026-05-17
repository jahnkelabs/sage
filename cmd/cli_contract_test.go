package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var sageTestBin string

func TestMain(m *testing.M) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		panic(err)
	}
	sageTestBin = filepath.Join(repoRoot, "testdata", "sage-testbin")
	build := exec.Command("go", "build", "-o", sageTestBin, ".")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		panic("build sage test binary: " + err.Error() + "\n" + string(out))
	}
	code := m.Run()
	_ = os.Remove(sageTestBin)
	os.Exit(code)
}

func runSage(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(sageTestBin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("run sage %v: %v", args, err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestCLIPassthrough_dryRun(t *testing.T) {
	dir := t.TempDir()
	composeFile, err := filepath.Abs(filepath.Join("..", "testdata", "merged-compose-example.yml"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		wantInLine []string
	}{
		{
			name:       "up detached",
			args:       []string{"--dry-run", "up", "-d"},
			wantInLine: []string{"up", "-d"},
		},
		{
			name:       "logs follow",
			args:       []string{"--dry-run", "logs", "-f", "api"},
			wantInLine: []string{"logs", "-f", "api"},
		},
		{
			name:       "up long detach flag",
			args:       []string{"--dry-run", "up", "--detach"},
			wantInLine: []string{"up", "--detach"},
		},
		{
			name:       "sage flags before verb",
			args:       []string{"--file", composeFile, "--dry-run", "ps"},
			wantInLine: []string{"-f", composeFile, "ps"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runSage(t, dir, tc.args...)
			if code != 0 {
				t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
			}
			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
			if !strings.Contains(stdout, "sage --dry-run would execute:") {
				t.Fatalf("missing dry-run header:\n%s", stdout)
			}
			for _, frag := range tc.wantInLine {
				if !strings.Contains(stdout, frag) {
					t.Fatalf("stdout missing %q:\n%s", frag, stdout)
				}
			}
		})
	}
}

func TestCLIAlias_dryRun(t *testing.T) {
	dir := t.TempDir()
	dump, err := filepath.Abs(filepath.Join("..", "testdata", "merged-compose-example.yml"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SAGE_COMPOSE_CONFIG_DUMP", dump)

	stdout, stderr, code := runSage(t, dir, "--dry-run", "migrate")
	if code != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "sage --dry-run would execute:") {
		t.Fatalf("missing dry-run header:\n%s", stdout)
	}
	for _, frag := range []string{
		"run",
		"--remove-orphans",
		"--rm",
		"api",
		"--profile",
		"dev",
		"bundle exec rake db:migrate",
	} {
		if !strings.Contains(stdout, frag) {
			t.Fatalf("stdout missing %q:\n%s", frag, stdout)
		}
	}
}

func TestCLIParseError_unknownFlag(t *testing.T) {
	dir := t.TempDir()
	_, stderr, code := runSage(t, dir, "--not-a-flag", "up")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "sage:") {
		t.Fatalf("stderr missing sage error prefix: %q", stderr)
	}
}

func TestCLIInstall_cursorRule(t *testing.T) {
	home := t.TempDir()
	stdout, stderr, code := runSage(t, "", "install", "--targets", "cursor", "--home", home)
	if code != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "installed") {
		t.Fatalf("stdout missing installed: %q", stdout)
	}

	rulePath := filepath.Join(home, ".cursor", "rules", "sage-compose.mdc")
	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, want := range []string{
		"alwaysApply: true",
		"sage aliases --json",
		"docker compose",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rule missing %q:\n%s", want, body)
		}
	}
}

func TestCLIInstall_skipWithoutForce(t *testing.T) {
	home := t.TempDir()
	_, _, code := runSage(t, "", "install", "--targets", "cursor", "--home", home)
	if code != 0 {
		t.Fatalf("first install exit %d", code)
	}
	stdout, stderr, code := runSage(t, "", "install", "--targets", "cursor", "--home", home)
	if code != 0 {
		t.Fatalf("second install exit %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "skipped") {
		t.Fatalf("stdout missing skipped: %q", stdout)
	}
}

func TestCLIInstall_forceOverwrite(t *testing.T) {
	home := t.TempDir()
	_, _, code := runSage(t, "", "install", "--targets", "cursor", "--home", home)
	if code != 0 {
		t.Fatalf("first install exit %d", code)
	}
	stdout, stderr, code := runSage(t, "", "install", "--targets", "cursor", "--home", home, "--force")
	if code != 0 {
		t.Fatalf("force install exit %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "installed") {
		t.Fatalf("stdout missing installed: %q", stdout)
	}
}

func TestCLIInstall_nonInteractiveRequiresTargets(t *testing.T) {
	home := t.TempDir()
	cmd := exec.Command(sageTestBin, "install", "--home", home)
	cmd.Stdin = strings.NewReader("")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	combined := outBuf.String() + errBuf.String()
	if !strings.Contains(combined, "--targets cursor") {
		t.Fatalf("missing usage hint: %q", combined)
	}
}
