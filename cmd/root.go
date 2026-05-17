package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var passthroughVerbs = []string{
	"build", "up", "start", "stop", "restart", "down", "ps", "logs", "run",
}

var (
	rootComposeFiles []string
	rootProject      string
	rootDryRun       bool

	aliasOnce      sync.Once
	aliasIndex     map[string]aliasEntry
	loadAliasesErr error
)

func resetAliasLoaderForTest() {
	aliasOnce = sync.Once{}
	aliasIndex = nil
	loadAliasesErr = nil
}

func ensureAliasesLoaded(ctx context.Context) {
	aliasOnce.Do(func() {
		ix, err := loadAliasIndex(ctx, rootComposeFiles, rootProject)
		if err != nil {
			aliasIndex = map[string]aliasEntry{}
			loadAliasesErr = err
			return
		}
		aliasIndex = ix
		loadAliasesErr = nil
	})
}

func isPassthrough(word string) bool {
	return slices.Contains(passthroughVerbs, strings.TrimSpace(word))
}

func composePrefix() []string {
	var out []string
	for _, f := range rootComposeFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		out = append(out, "-f", f)
	}
	if strings.TrimSpace(rootProject) != "" {
		out = append(out, "-p", strings.TrimSpace(rootProject))
	}
	return out
}

func warnSubst(msg string) {
	fmt.Fprintf(os.Stderr, "sage: warning: %s\n", msg)
}

func quoteArgv(argv []string) string {
	var b strings.Builder
	for i, a := range argv {
		if i > 0 {
			b.WriteByte(' ')
		}
		if strings.ContainsAny(a, " \t\n\"'`$\\") {
			b.WriteString(fmt.Sprintf("%q", a))
			continue
		}
		b.WriteString(a)
	}
	return b.String()
}

func printDryRun(composeArgs []string) {
	full := composeInvocationPreview(composeArgs)
	fmt.Fprintf(os.Stdout, "sage --dry-run would execute:\n  %s\n", quoteArgv(full))
}

func forwardCompose(ctx context.Context, args []string) error {
	full := append(composePrefix(), args...)
	if rootDryRun {
		printDryRun(full)
		return nil
	}
	err, code := dockerComposeRun(ctx, full)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sage: %v\n", err)
	}
	os.Exit(code)
	return nil
}

func runAlias(ctx context.Context, ae aliasEntry, extra []string) error {
	expanded := expandEnv(ae.RawCmd, warnSubst)
	parts := strings.Fields(expanded)

	base := composePrefix()
	for _, p := range ae.Profiles {
		base = append(base, "--profile", p)
	}

	runArgs := append(base, "run")
	if !skipTTYForCompose() {
		runArgs = append(runArgs, "-it")
	}
	runArgs = append(runArgs, "--rm", "--remove-orphans", ae.Service)
	runArgs = append(runArgs, parts...)
	runArgs = append(runArgs, extra...)

	if rootDryRun {
		printDryRun(runArgs)
		return nil
	}

	err, code := dockerComposeRun(ctx, runArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sage: %v\n", err)
	}
	os.Exit(code)
	return nil
}

func appendAliasesHelp(out io.Writer, loadedErr error) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Aliases (sage.alias.* labels from merged compose config):")
	if loadedErr != nil {
		fmt.Fprintf(out, "  unavailable (%v)\n", loadedErr)
		return
	}
	if len(aliasIndex) == 0 {
		fmt.Fprintln(out, "  (none discovered)")
		return
	}

	type row struct {
		alias, svc, cmd string
	}
	var rows []row
	for k, ae := range aliasIndex {
		rows = append(rows, row{k, ae.Service, ae.RawCmd})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].alias < rows[j].alias })

	const maxCmd = 56
	for _, r := range rows {
		cmd := r.cmd
		if len(cmd) > maxCmd {
			cmd = cmd[:maxCmd-3] + "..."
		}
		fmt.Fprintf(out, "  %-14s %-12s %s\n", r.alias, r.svc, cmd)
	}
}

// Execute starts the sage CLI.
func Execute(version, commit, date string) {
	rootCmd := newRootCmd(version, commit, date)
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "sage: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sage [flags] <compose verb | alias> [args...]",
		Short: "Docker Compose helper with label-defined command aliases",
		Long: strings.TrimSpace(`
sage forwards familiar docker compose verbs unchanged, and expands shortcuts declared as
compose service labels of the form sage.alias.<name>.

Examples:
  sage up -d
  sage migrate
  sage --dry-run test`),
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			first := args[0]
			rest := args[1:]

			if isPassthrough(first) || strings.HasPrefix(first, "-") {
				return forwardCompose(ctx, args)
			}

			ensureAliasesLoaded(ctx)
			if ae, ok := aliasIndex[first]; ok {
				return runAlias(ctx, ae, rest)
			}

			return forwardCompose(ctx, args)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			ensureAliasesLoaded(ctx)

			var matches []string
			for _, v := range passthroughVerbs {
				if strings.HasPrefix(v, toComplete) {
					matches = append(matches, v+"\tdocker compose verb")
				}
			}
			for alias := range aliasIndex {
				if strings.HasPrefix(alias, toComplete) {
					matches = append(matches, alias+"\tcompose alias")
				}
			}
			sort.Strings(matches)
			return matches, cobra.ShellCompDirectiveDefault
		},
	}

	// Use --file instead of -f so compose verbs like `logs -f` keep docker's -f flag.
	rootCmd.PersistentFlags().StringSliceVar(&rootComposeFiles, "file", nil, "Compose file paths (repeatable)")
	rootCmd.PersistentFlags().StringVarP(&rootProject, "project-name", "p", "", "Compose project name")
	rootCmd.PersistentFlags().BoolVar(&rootDryRun, "dry-run", false, "Print docker compose invocation instead of running")

	rootCmd.Version = formatVersion(version, commit, date)
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		ensureAliasesLoaded(ctx)

		out := cmd.OutOrStdout()
		fmt.Fprint(out, strings.TrimRight(cmd.UsageString(), "\n"))
		appendAliasesHelp(out, loadAliasesErr)
		fmt.Fprintln(out)
	})

	rootCmd.InitDefaultHelpCmd()

	aliasesCmd := &cobra.Command{
		Use:   "aliases",
		Short: "List discovered sage.alias.* aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			ensureAliasesLoaded(ctx)
			if loadAliasesErr != nil {
				return fmt.Errorf("aliases unavailable: %w", loadAliasesErr)
			}
			asJSON, _ := cmd.Flags().GetBool("json")
			type row struct {
				Alias   string `json:"alias"`
				Service string `json:"service"`
				Command string `json:"command"`
			}
			var rows []row
			for alias, ae := range aliasIndex {
				rows = append(rows, row{Alias: alias, Service: ae.Service, Command: ae.RawCmd})
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].Alias < rows[j].Alias })

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-18s %-14s %s\n", "ALIAS", "SERVICE", "COMMAND")
			fmt.Fprintf(cmd.OutOrStdout(), "%-18s %-14s %s\n", strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 10))
			const maxCmd = 96
			for _, r := range rows {
				cmdStr := r.Command
				if len(cmdStr) > maxCmd {
					cmdStr = cmdStr[:maxCmd-3] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-18s %-14s %s\n", r.Alias, r.Service, cmdStr)
			}
			return nil
		},
	}
	aliasesCmd.Flags().Bool("json", false, "Emit aliases as JSON")

	completionCmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish]",
		Short:                 "Generate shell completion script",
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return fmt.Errorf("unsupported shell %q (use bash, zsh, or fish)", args[0])
			}
		},
	}

	rootCmd.AddCommand(aliasesCmd, completionCmd, newInstallCmd())

	// Compose flags (e.g. up -d, logs -f) must follow the verb; only sage flags before it.
	rootCmd.PersistentFlags().SetInterspersed(false)
	rootCmd.Flags().SetInterspersed(false)

	return rootCmd
}

func formatVersion(version, commit, date string) string {
	c := commit
	if len(c) > 7 {
		c = c[:7]
	}
	return fmt.Sprintf("%s (commit %s, %s)", version, c, date)
}
