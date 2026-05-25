package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func dim(s string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", s)
}

// aliasCol returns the alias padded to 5 chars and dimmed, for consistent column alignment.
func aliasCol(alias string) string {
	return dim(fmt.Sprintf("%-5s", alias))
}

func newProjectsCmd() *cobra.Command {
	var open, local, remote, set bool
	var mode string

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List projects from tabs, local dirs, and/or GitHub repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			var include []string
			for _, f := range []string{"open", "local", "remote", "set"} {
				if v, _ := flags.GetBool(f); v {
					include = append(include, "--"+f)
				}
			}
			if len(include) == 0 {
				include = []string{"--open", "--local"}
			}
			items, err := runProjects(false, mode, include...)
			if err != nil {
				return err
			}
			if len(items) > 0 {
				fmt.Print(strings.Join(items, "\n") + "\n")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&open, "open", false, "Include open project")
	cmd.Flags().BoolVar(&local, "local", false, "Include local projects")
	cmd.Flags().BoolVar(&remote, "remote", false, "Include remote projects")
	cmd.Flags().BoolVar(&set, "set", false, "Include configured sets")
	cmd.Flags().StringVar(&mode, "mode", "", "Prefix each item with this mode (open|close) for use with fzf --with-nth")

	return cmd
}

func runProjects(refresh bool, mode string, flags ...string) ([]string, error) {
	var items []string

	prefix := ""
	if mode != "" {
		prefix = mode + "\t"
	}

	for _, flag := range flags {
		switch flag {
		case "--open":
			tabs, err := term.ListTabs()
			if err != nil {
				return nil, err
			}
			for _, t := range tabs {
				if t.Title != "" {
					name := nameFor(t.Title)
					if name != t.Title {
						items = append(items, prefix+fmt.Sprintf("%s  %s", aliasCol(t.Title), name))
					} else {
						items = append(items, prefix+fmt.Sprintf("%s  %s", aliasCol(""), t.Title))
					}
				}
			}
		case "--local":
			for _, dir := range viper.GetStringSlice("dirs") {
				if strings.HasSuffix(dir, "/") {
					dir = expandHome(dir)
					entries, err := os.ReadDir(dir)
					if err != nil {
						return nil, fmt.Errorf("could not read %s: %w", dir, err)
					}
					for _, e := range entries {
						if e.IsDir() {
							name := e.Name()
							path := collapseHome(filepath.Join(dir, name))
							alias := aliasFor(name)
							items = append(items, prefix+fmt.Sprintf("%s  %s", aliasCol(alias), path))
						}
					}
				} else {
					name := filepath.Base(expandHome(dir))
					path := collapseHome(expandHome(dir))
					alias := aliasFor(name)
					items = append(items, prefix+fmt.Sprintf("%s  %s", aliasCol(alias), path))
				}
			}
		case "--remote":
			for _, r := range viper.GetStringSlice("github") {
				repos, err := cachedListRepos(r, refresh)
				if err != nil {
					return nil, err
				}
				for _, repo := range repos {
					items = append(items, prefix+fmt.Sprintf("%-30s %s", repo.Name, dim(repo.SSHUrl)))
				}
			}
		case "--set":
			for _, b := range getSets() {
				items = append(items, prefix+fmt.Sprintf("%-30s %s", b.Name, dim(strings.Join(b.Projects, ", "))))
			}
		}
	}

	return items, nil
}
