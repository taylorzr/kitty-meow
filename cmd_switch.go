package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
				{viper.GetString("keybindings.history"), "history", fmt.Sprintf("%s history --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.set"), "set", fmt.Sprintf("%s sets --mode=open 2>&1", exe)},
				{viper.GetString("keybindings.default"), "default", defaultCmd},
				{viper.GetString("keybindings.close"), "close", fmt.Sprintf("%s projects --open --mode=close 2>&1", exe)},
			})

			fzf := exec.Command(viper.GetString("fzf"),
				"--prompt=🐈 switch > ",
				"--header="+header,
				"--bind="+binds,
				// "--bind=tab:toggle",
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
					title := strings.Fields(ansiRe.ReplaceAllString(content, ""))[0]
					if err := term.CloseTab(title); err != nil {
						fmt.Fprintf(os.Stderr, "failed to close tab %q: %v\n", title, err)
					}
				} else {
					if err := handleSelection(content, cloneDir); err != nil {
						return err
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

// Selection formats:
//   - "reponame git@github.com:owner/repo.git" → remote: clone then open
//   - "~/code/project" or "/abs/path/project"  → local dir: open as new tab
//   - "mytab"                                   → open tab: focus it
func handleSelection(selection string, cloneDir string) error {
	// Strip ANSI codes
	selection = strings.TrimSpace(ansiRe.ReplaceAllString(selection, ""))

	// Remote: "name<padding>sshUrl" — detect URL before double-space truncation,
	// since %-30s padding creates double spaces that would strip the URL away.
	fields := strings.Fields(selection)
	if len(fields) >= 2 {
		last := fields[len(fields)-1]
		if strings.HasPrefix(last, "git@") || strings.HasPrefix(last, "https://") {
			name := fields[0]
			sshURL := last
			recordHistory(name)
			tabs, err := term.ListTabs()
			if err == nil {
				for _, t := range tabs {
					if t.Title == name {
						return term.FocusTab(name)
					}
				}
			}
			dest := expandHome(strings.TrimSuffix(cloneDir, "/") + "/" + name)
			if _, err := os.Stat(dest); os.IsNotExist(err) {
				fmt.Printf("cloning %s into %s...\n", name, dest)
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
			return term.NewTab(name, dest)
		}
	}

	// Strip trailing annotation for history entries ("name   2024-01-01T12:00:00Z" → "name")
	if i := strings.Index(selection, "  "); i != -1 {
		selection = strings.TrimSpace(selection[:i])
	}

	// Local dir: path starting with ~ or /
	if strings.HasPrefix(selection, "~/") || strings.HasPrefix(selection, "/") {
		path := expandHome(selection)
		name := path[strings.LastIndex(path, "/")+1:]
		recordHistory(name)
		tabs, err := term.ListTabs()
		if err == nil {
			for _, t := range tabs {
				if t.Title == name {
					return term.FocusTab(name)
				}
			}
		}
		return term.NewTab(name, path)
	}

	// Alias: use alias as tab title, resolve real name for path lookup
	if aliases := viper.GetStringMapString("aliases"); len(aliases) > 0 {
		if realName, ok := aliases[selection]; ok {
			tabs, err := term.ListTabs()
			if err == nil {
				for _, t := range tabs {
					if t.Title == selection || t.Title == realName {
						recordHistory(realName)
						return term.FocusTab(t.Title)
					}
				}
			}
			if path := findLocalPath(realName); path != "" {
				recordHistory(realName)
				return term.NewTab(selection, expandHome(path))
			}
		}
	}

	// Set: expand to individual projects
	for _, b := range getSets() {
		if b.Name == selection {
			for _, project := range b.Projects {
				if err := handleSelection(project, cloneDir); err != nil {
					fmt.Fprintf(os.Stderr, "set: failed to open %q: %v\n", project, err)
				}
			}
			return nil
		}
	}

	// Open tab: focus if open, otherwise find local dir and open it
	if path := findLocalPath(selection); path != "" {
		return handleSelection(expandHome(path), cloneDir)
	}
	recordHistory(selection)
	return term.FocusTab(selection)
}

func recordHistory(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "kitty", "meow")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "history")

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	lines = append(lines, fmt.Sprintf("%s %s", name, time.Now().Format(time.RFC3339)))

	if len(lines) > viper.GetInt("history_size") {
		lines = lines[len(lines)-viper.GetInt("history_size"):]
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
