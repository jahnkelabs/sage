package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// composeInvocationPreview returns argv as printed by dry-run (binary + compose args).
func composeInvocationPreview(composeArgs []string) []string {
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return append([]string{"docker-compose"}, composeArgs...)
	}
	return append([]string{"docker", "compose"}, composeArgs...)
}

// dockerComposeCmd prefers standalone docker-compose when present; otherwise docker compose.
func dockerComposeCmd(ctx context.Context, args ...string) *exec.Cmd {
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return exec.CommandContext(ctx, "docker-compose", args...)
	}
	return exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
}

func runWithSignalForwarding(proc *exec.Cmd) (error, int) {
	sigCh := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := proc.Start(); err != nil {
		signal.Stop(sigCh)
		return err, 1
	}

	go func() {
		for {
			select {
			case sig := <-sigCh:
				_ = proc.Process.Signal(sig)
			case <-done:
				return
			}
		}
	}()

	err := proc.Wait()
	signal.Stop(sigCh)
	close(done)
	if proc.ProcessState != nil {
		return err, proc.ProcessState.ExitCode()
	}
	return err, 1
}

// dockerComposeRun forwards stdin/stdout/stderr from the OS.
func dockerComposeRun(ctx context.Context, composeArgs []string) (error, int) {
	cmd := dockerComposeCmd(ctx, composeArgs...)
	var stdin io.Reader = os.Stdin
	if skipTTYForCompose() {
		stdin = strings.NewReader("")
	}
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return runWithSignalForwarding(cmd)
}

// skipTTYForCompose disables interactive stdin when SAGE_NO_TTY is set (CI/agents).
func skipTTYForCompose() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SAGE_NO_TTY")))
	return v == "1" || v == "true" || v == "yes"
}

// dockerComposeOutput captures stdout for config merges.
func dockerComposeOutput(ctx context.Context, composeArgs []string) ([]byte, error) {
	cmd := dockerComposeCmd(ctx, composeArgs...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
