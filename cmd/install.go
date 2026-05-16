package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jahnkelabs/sage/integrations"
	"github.com/spf13/cobra"
)

const cursorRuleName = "sage-compose.mdc"

var (
	installTargets []string
	installHome    string
	installForce   bool
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Sage agent integrations into your user config",
		Long: strings.TrimSpace(`
Install bundled agent integrations into your home directory.

Currently installs a Cursor user rule (~/.cursor/rules/sage-compose.mdc) that instructs
agents to prefer sage for Docker Compose and to discover sage.alias.* shortcuts.

Re-run after upgrading sage to refresh the rule. Confirm in Cursor Settings → Rules
that sage-compose shows Always Apply.`),
		RunE: runInstall,
	}
	cmd.Flags().StringSliceVar(&installTargets, "targets", nil, "Non-interactive targets (cursor)")
	cmd.Flags().StringVar(&installHome, "home", "", "Home directory override (default: user home; for tests)")
	cmd.Flags().BoolVar(&installForce, "force", false, "Overwrite existing files")
	return cmd
}

func runInstall(cmd *cobra.Command, _ []string) error {
	home, err := resolveInstallHome(installHome)
	if err != nil {
		return err
	}

	targets := installTargets
	if len(targets) == 0 {
		if !isInteractiveTerminal(cmd.InOrStdin()) {
			return fmt.Errorf("non-interactive shell: pass --targets cursor (e.g. sage install --targets cursor)")
		}
		dest := filepath.Join(home, ".cursor", "rules", cursorRuleName)
		if !confirmInstall(cmd.InOrStdin(), cmd.OutOrStdout(), dest) {
			fmt.Fprintln(cmd.OutOrStdout(), "skipped")
			return nil
		}
		targets = []string{"cursor"}
	}

	var hadErr bool
	for _, t := range targets {
		t = strings.TrimSpace(strings.ToLower(t))
		switch t {
		case "cursor":
			result, err := installCursorRule(home, installForce)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "sage: cursor: %v\n", err)
				hadErr = true
				continue
			}
			dest := filepath.Join(home, ".cursor", "rules", cursorRuleName)
			switch result {
			case installSkipped:
				fmt.Fprintf(cmd.OutOrStdout(), "skipped %s (already exists; use --force to overwrite)\n", dest)
			case installWritten:
				fmt.Fprintf(cmd.OutOrStdout(), "installed %s\n", dest)
				fmt.Fprintln(cmd.OutOrStdout(), "confirm in Cursor Settings → Rules that sage-compose shows Always Apply")
			}
		default:
			fmt.Fprintf(cmd.ErrOrStderr(), "sage: unknown target %q (supported: cursor)\n", t)
			hadErr = true
		}
	}
	if hadErr {
		return fmt.Errorf("install failed")
	}
	return nil
}

type installResult int

const (
	installWritten installResult = iota
	installSkipped
)

func resolveInstallHome(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return filepath.Clean(override), nil
	}
	return os.UserHomeDir()
}

func isInteractiveTerminal(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func confirmInstall(in io.Reader, out io.Writer, dest string) bool {
	fmt.Fprintf(out, "Install sage-compose rule to %s? [Y/n] ", dest)
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		return true
	}
	line := strings.TrimSpace(strings.ToLower(sc.Text()))
	return line == "" || line == "y" || line == "yes"
}

func installCursorRule(home string, force bool) (installResult, error) {
	data, err := integrations.CursorRules.ReadFile("cursor/rules/" + cursorRuleName)
	if err != nil {
		return installSkipped, fmt.Errorf("read embedded rule: %w", err)
	}

	destDir := filepath.Join(home, ".cursor", "rules")
	dest := filepath.Join(destDir, cursorRuleName)

	if _, err := os.Stat(dest); err == nil && !force {
		return installSkipped, nil
	} else if err != nil && !os.IsNotExist(err) {
		return installSkipped, err
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return installSkipped, fmt.Errorf("create rules dir: %w", err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return installSkipped, fmt.Errorf("write rule: %w", err)
	}
	return installWritten, nil
}
