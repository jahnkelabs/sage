package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const aliasLabelPrefix = "sage.alias."

type aliasEntry struct {
	Service string
	RawCmd  string
	Profiles []string
}

type composeModel struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Labels   any      `yaml:"labels"`
	Profiles []string `yaml:"profiles"`
}

// mergeComposeYAML shells out to docker compose config with every discovered profile included.
func mergeComposeYAML(ctx context.Context, prefixArgs []string) ([]byte, error) {
	if dump := strings.TrimSpace(os.Getenv("SAGE_COMPOSE_CONFIG_DUMP")); dump != "" {
		return os.ReadFile(dump)
	}

	profilesCmd := dockerComposeCmd(ctx, append(append([]string{}, prefixArgs...), "config", "--profiles")...)
	profilesCmd.Stderr = os.Stderr
	profilesOut, profilesErr := profilesCmd.Output()

	configArgs := append([]string{}, prefixArgs...)
	if profilesErr == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(profilesOut)), "\n") {
			p := strings.TrimSpace(line)
			if p != "" {
				configArgs = append(configArgs, "--profile", p)
			}
		}
	}
	configArgs = append(configArgs, "config")

	data, err := dockerComposeOutput(ctx, configArgs)
	if err != nil {
		return nil, fmt.Errorf("docker compose config: %w", err)
	}
	return data, nil
}

func parseAliasesFromYAML(raw []byte) (map[string]aliasEntry, error) {
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	var merged composeModel
	if err := dec.Decode(&merged); err != nil {
		return nil, fmt.Errorf("parse compose YAML: %w", err)
	}

	out := make(map[string]aliasEntry)
	for svcName, def := range merged.Services {
		if def.Labels == nil {
			continue
		}
		labels, ok := def.Labels.(map[string]interface{})
		if !ok {
			continue
		}
		for k, v := range labels {
			strVal, ok := v.(string)
			if !ok {
				continue
			}
			if !strings.HasPrefix(k, aliasLabelPrefix) {
				continue
			}
			aliasName := strings.TrimPrefix(k, aliasLabelPrefix)
			if aliasName == "" {
				continue
			}
			if prev, dup := out[aliasName]; dup {
				return nil, fmt.Errorf("duplicate alias %q on services %q and %q", aliasName, prev.Service, svcName)
			}
			out[aliasName] = aliasEntry{
				Service:  svcName,
				RawCmd:   strVal,
				Profiles: append([]string(nil), def.Profiles...),
			}
		}
	}
	if len(out) == 0 {
		return out, nil
	}
	return out, nil
}

func loadAliasIndex(ctx context.Context, composeFiles []string, project string) (map[string]aliasEntry, error) {
	var prefix []string
	for _, f := range composeFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		prefix = append(prefix, "-f", f)
	}
	if strings.TrimSpace(project) != "" {
		prefix = append(prefix, "-p", project)
	}

	raw, err := mergeComposeYAML(ctx, prefix)
	if err != nil {
		return nil, err
	}
	return parseAliasesFromYAML(raw)
}
