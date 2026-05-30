package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func defaultSourceFlags() []string {
	sources := viper.GetStringSlice("default_sources")
	var flags []string
	for _, s := range sources {
		flags = append(flags, "--"+s)
	}
	return flags
}

func newSwitchCmd() *cobra.Command {
	viper.SetDefault("keybindings.remote", "ctrl-g")
	viper.SetDefault("keybindings.local", "ctrl-p")
	viper.SetDefault("keybindings.open", "ctrl-o")
	viper.SetDefault("keybindings.set", "ctrl-s")
	viper.SetDefault("keybindings.history", "ctrl-r")
	viper.SetDefault("keybindings.default", "ctrl-a")
	viper.SetDefault("keybindings.close", "ctrl-x")
	viper.SetDefault("default_sources", []string{"open", "local"})
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch project via fzf",
		RunE: func(cmd *cobra.Command, args []string) error {
			exe, err := os.Executable()
			if err != nil {
				return err
			}

			defaultFlags := defaultSourceFlags()
			initial, err := runProjects(false, "open", defaultFlags...)
			if err != nil {
				return err
			}

			defaultCmd := fmt.Sprintf("%s projects %s --mode=open 2>&1", exe, strings.Join(defaultFlags, " "))
			binds, header := bindsAndHeader("🐈", []bind{
				{viper.GetString("keybindings.remote"), "remote", fmt.Sprintf("%s projects --remote --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.local"), "local", fmt.Sprintf("%s projects --local --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.open"), "open", fmt.Sprintf("%s projects --open --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.history"), "history", fmt.Sprintf("%s projects --history --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.set"), "set", fmt.Sprintf("%s projects --set --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.default"), "default", defaultCmd},
				{viper.GetString("keybindings.close"), "close", fmt.Sprintf("%s projects --open --mode=close 2>&1", exe)},
			})

			fzf := exec.Command(viper.GetString("fzf"),
				"--prompt=🐈 switch > ",
				"--header="+header,
				"--bind="+binds,
				"--reverse",
				"--ansi",
				"--multi",
				"--delimiter=\t",
				"--with-nth=2..",
			)

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Println(strings.Join(fzf.Args, " "))
				fmt.Println("--- stdin ---")
				fmt.Println(strings.Join(initial, "\n"))
				return nil
			}

			var out strings.Builder
			fzf.Stdin = strings.NewReader(strings.Join(initial, "\n"))
			fzf.Stdout = &out
			fzf.Stderr = os.Stderr
			if err := fzf.Run(); err != nil {
				// NOTE: ctrl-c treated as normal exit
				if err.Error() == "exit status 130" {
					return nil
				}

				fmt.Fprintf(os.Stderr, "fzf failed: %v\nPress Enter to continue...", err)
				tty, _ := os.Open("/dev/tty")
				if tty != nil {
					bufio.NewReader(tty).ReadBytes('\n')
					tty.Close()
				}
			}

			selection := strings.TrimSpace(out.String())
			if selection == "" {
				return nil
			}

			cloneDir := ""
			dirs := viper.GetStringSlice("dirs")
			if len(dirs) > 0 {
				cloneDir = dirs[0]
			}
			for _, line := range strings.Split(selection, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				mode, content, _ := strings.Cut(line, "\t")
				if mode == "close" {
					// content is "aliasCol\tname[\trelTime]" — parts[1] is the display name
					parts := strings.Split(ansiRe.ReplaceAllString(content, ""), "\t")
					var title string
					if len(parts) >= 2 {
						title = strings.TrimSpace(parts[1])
					} else {
						title = strings.TrimSpace(parts[0])
					}
					// If this name has an alias, the tab is titled with the alias key
					if alias := aliasFor(title); alias != "" {
						title = alias
					}
					if err := term.CloseTab(title); err != nil {
						fmt.Fprintf(os.Stderr, "failed to close tab %q: %v\n", title, err)
					}
				} else {
					if err := handleSelection(content, cloneDir); err != nil {
						fmt.Fprintf(os.Stderr, "failed to open %q: %v\n", content, err)
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().Bool("dry-run", false, "Print the fzf command and input instead of running it")
	return cmd
}

func findLocalPath(name string) string {
	for _, dir := range viper.GetStringSlice("dirs") {
		if strings.HasSuffix(dir, "/") {
			candidate := filepath.Join(expandHome(dir), name)
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return collapseHome(candidate)
			}
		}
	}
	return ""
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Selection formats (tab-delimited columns, ANSI codes present):
//   - "aliasCol\t~/code/project"                    → local dir: open as new tab
//   - "aliasCol\ttabTitle"                           → open tab: focus it
//   - "aliasCol\trepoName\tgit@github.com:owner/repo" → remote: clone then open
//   - "aliasCol\tsetName\tdim(projects)"             → set: expand to projects
//   - "aliasCol\tprojectName\tdim(timestamp)"        → history entry
//
// All display items use aliasCol (parts[0]) as the first column; the identifier is parts[1].
// When called recursively (e.g. from set expansion), selection is a bare project name or path.
func handleSelection(selection string, cloneDir string) error {
	// Strip ANSI codes, then split on tabs. Do NOT TrimSpace the whole string first —
	// the leading alias column is spaces that would otherwise eat the first tab separator.
	raw := ansiRe.ReplaceAllString(selection, "")
	parts := strings.Split(raw, "\t")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	// Resolve the identifier: parts[1] when there are 2+ columns (alias col + identifier),
	// otherwise parts[0] (bare recursive call).
	identifier := parts[0]
	if len(parts) >= 2 {
		identifier = parts[1]
	}

	// Remote: last field is an SSH or HTTPS URL (format: "aliasCol\trepoName\tgit@...")
	last := parts[len(parts)-1]
	if strings.HasPrefix(last, "git@") || strings.HasPrefix(last, "https://") {
		sshURL := last
		recordHistory(identifier)
		tabs, err := term.ListTabs()
		if err == nil {
			for _, t := range tabs {
				if t.Title == identifier {
					return term.FocusTab(identifier)
				}
			}
		}
		dest := expandHome(strings.TrimSuffix(cloneDir, "/") + "/" + identifier)
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			fmt.Printf("cloning %s into %s...\n", identifier, dest)
			cloneCmd := exec.Command("git", "clone", sshURL, dest)
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			if err := cloneCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git clone failed: %v\nPress Enter to continue...", err)
				tty, _ := os.Open("/dev/tty")
				if tty != nil {
					bufio.NewReader(tty).ReadBytes('\n')
					tty.Close()
				}
				return nil
			}
		}
		return term.NewTab(identifier, dest)
	}

	// Local dir: identifier is a path starting with ~ or /
	if strings.HasPrefix(identifier, "~/") || strings.HasPrefix(identifier, "/") {
		path := expandHome(identifier)
		name := path[strings.LastIndex(path, "/")+1:]
		recordHistory(name)
		// Tab may be open under the alias name rather than the real name.
		alias := aliasFor(name)
		tabs, err := term.ListTabs()
		if err == nil {
			for _, t := range tabs {
				if t.Title == name || (alias != "" && t.Title == alias) {
					return term.FocusTab(t.Title)
				}
			}
		}
		// Open a new tab; prefer alias as the tab title when one is configured.
		tabTitle := name
		if alias != "" {
			tabTitle = alias
		}
		return term.NewTab(tabTitle, path)
	}

	// Alias: identifier matches an alias key → use alias as tab title, resolve real name for path
	if aliases := viper.GetStringMapString("aliases"); len(aliases) > 0 {
		if realName, ok := aliases[identifier]; ok {
			tabs, err := term.ListTabs()
			if err == nil {
				for _, t := range tabs {
					if t.Title == identifier || t.Title == realName {
						recordHistory(realName)
						return term.FocusTab(t.Title)
					}
				}
			}
			if path := findLocalPath(realName); path != "" {
				recordHistory(realName)
				return term.NewTab(identifier, expandHome(path))
			}
		}
	}

	// Set: expand to individual projects
	for _, b := range getSets() {
		if b.Name == identifier {
			for _, project := range b.Projects {
				if err := handleSelection(project, cloneDir); err != nil {
					fmt.Fprintf(os.Stderr, "set: failed to open %q: %v\n", project, err)
				}
			}
			return nil
		}
	}

	// Open tab: focus if open, otherwise find local dir and open it
	if path := findLocalPath(identifier); path != "" {
		return handleSelection(expandHome(path), cloneDir)
	}
	recordHistory(identifier)
	return term.FocusTab(identifier)
}
