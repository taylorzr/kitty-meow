package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func dim(s string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", s)
}

// aliasCol returns the alias padded to alias_spacing chars and dimmed.
func aliasCol(alias string) string {
	w := viper.GetInt("alias_spacing")
	return dim(fmt.Sprintf("%-*s", w, alias))
}

// nameCol pads a project name to name_spacing chars for annotation alignment.
func nameCol(name string) string {
	w := viper.GetInt("name_spacing")
	if len(name) >= w {
		return name + "  "
	}
	return fmt.Sprintf("%-*s", w, name)
}

func newProjectsCmd() *cobra.Command {
	var open, local, remote, set, history bool
	var mode string

	cmd := &cobra.Command{
		Use:   "projects",
		Short: "List projects from tabs, local dirs, and/or GitHub repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			var include []string
			for _, f := range []string{"open", "local", "remote", "set", "history"} {
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
	cmd.Flags().BoolVar(&open, "open", false, "Include open tabs")
	cmd.Flags().BoolVar(&local, "local", false, "Include local projects")
	cmd.Flags().BoolVar(&remote, "remote", false, "Include remote projects")
	cmd.Flags().BoolVar(&set, "set", false, "Include configured sets")
	cmd.Flags().BoolVar(&history, "history", false, "Include recently visited projects from history")
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
			var lastViews map[string]time.Time
			var now time.Time
			if mode == "close" {
				lastViews = readLastViews()
				now = time.Now()
			}
			for _, t := range tabs {
				if t.Title != "" {
					name := nameFor(t.Title)
					var alias string
					if name != t.Title {
						alias = t.Title
					}
					var entry string
					if lastViews != nil {
						if last, ok := lastViews[name]; ok {
							rel := dim(relativeTime(now.Sub(last)))
							entry = prefix + aliasCol(alias) + "\t" + nameCol(name) + "\t" + rel
						}
					}
					if entry == "" {
						entry = prefix + aliasCol(alias) + "\t" + name
					}
					items = append(items, entry)
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
							items = append(items, prefix+aliasCol(alias)+"\t"+path)
						}
					}
				} else {
					name := filepath.Base(expandHome(dir))
					path := collapseHome(expandHome(dir))
					alias := aliasFor(name)
					items = append(items, prefix+aliasCol(alias)+"\t"+path)
				}
			}
		case "--remote":
			for _, r := range viper.GetStringSlice("github") {
				repos, err := cachedListRepos(r, refresh)
				if err != nil {
					return nil, err
				}
				for _, repo := range repos {
					items = append(items, prefix+aliasCol("")+"\t"+nameCol(repo.Name)+"\t"+dim(repo.SSHUrl))
				}
			}
		case "--set":
			for _, b := range getSets() {
				items = append(items, prefix+aliasCol("")+"\t"+nameCol(b.Name)+"\t"+dim(strings.Join(b.Projects, ", ")))
			}
		case "--history":
			histItems, err := readHistory()
			if err != nil {
				return nil, err
			}
			for _, item := range histItems {
				items = append(items, prefix+item)
			}
		}
	}

	return items, nil
}

type setConfig struct {
	Name     string   `mapstructure:"name"`
	Projects []string `mapstructure:"projects"`
}

func getSets() []setConfig {
	var sets []setConfig
	viper.UnmarshalKey("sets", &sets)
	return sets
}
