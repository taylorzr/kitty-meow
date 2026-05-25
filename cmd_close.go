package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCloseCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close old tabs via fzf",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug, _ := cmd.Flags().GetBool("debug")

			tabs, err := term.ListTabs()
			if err != nil {
				return err
			}

			lastViews := map[string]time.Time{}
			home, _ := os.UserHomeDir()
			histPath := filepath.Join(home, ".config", "kitty", "meow", "history")
			if data, err := os.ReadFile(histPath); err == nil {
				for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
					var name, tsStr string
					if tab := strings.IndexByte(line, '\t'); tab != -1 {
						name = line[:tab]
						tsStr = line[tab+1:]
					} else {
						// Legacy format: "name timestamp"
						fields := strings.Fields(line)
						if len(fields) >= 2 {
							name = strings.Join(fields[:len(fields)-1], " ")
							tsStr = fields[len(fields)-1]
						} else if len(fields) == 1 {
							name = fields[0]
						}
					}
					if name != "" && tsStr != "" {
						if t, err := parseTimestamp(tsStr); err == nil {
							lastViews[name] = t
						}
					}
				}
			}

			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			cutoff := today.AddDate(0, 0, -days)

			if debug {
				fmt.Fprintf(os.Stderr, "cutoff: %s\n", cutoff.Format(time.RFC3339))
				fmt.Fprintf(os.Stderr, "open tabs (%d):\n", len(tabs))
				for _, t := range tabs {
					last, seen := lastViews[nameFor(t.Title)]
					if seen {
						fmt.Fprintf(os.Stderr, "  %q last=%s old=%v\n", t.Title, last.Format(time.RFC3339), last.Before(cutoff))
					} else {
						fmt.Fprintf(os.Stderr, "  %q (not in history → old)\n", t.Title)
					}
				}
			}

			var allTabs, oldTabs []string
		for _, t := range tabs {
				if t.Title == "" {
					continue
				}
				name := nameFor(t.Title)
				var entry string
				if name != t.Title {
					entry = aliasCol(t.Title) + "\t" + name
				} else {
					entry = aliasCol("") + "\t" + t.Title
				}
				allTabs = append(allTabs, entry)
				last, seen := lastViews[name]
				if !seen || last.Before(cutoff) {
					oldTabs = append(oldTabs, entry)
				}
			}

			binds, header := bindsAndHeader("🐈💀", []bind{
				{key: "ctrl-o", prompt: "old", command: fmt.Sprintf("printf '%s'", strings.Join(oldTabs, "\\n"))},
				{key: "ctrl-a", prompt: "all", command: fmt.Sprintf("printf '%s'", strings.Join(allTabs, "\\n"))},
			})

			// TODO: abstract out the fzf invocation since it's basically the same in all commands
			fzf := exec.Command(viper.GetString("fzf"),
				"--prompt=🐈💀 close > ",
				"--header="+header,
				"--bind="+binds,
				"--reverse",
				"--multi",
				"--ansi",
				"--delimiter=\t",
				"--with-nth=2..",
			)
			fzf.Stdin = strings.NewReader(strings.Join(oldTabs, "\n"))
			var out strings.Builder
			fzf.Stdout = &out
			fzf.Stderr = os.Stderr

			if err := fzf.Run(); err != nil {
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

			for title := range strings.SplitSeq(selection, "\n") {
				title = strings.TrimSpace(title)
				if title == "" {
					continue
				}
				// Entry is "aliasCol\ttabTitle" — take the tab-delimited title field
				parts := strings.SplitN(ansiRe.ReplaceAllString(title, ""), "\t", 2)
				tabTitle := strings.TrimSpace(parts[len(parts)-1])
				if err := term.CloseTab(tabTitle); err != nil {
					fmt.Fprintf(os.Stderr, "failed to close tab %q: %v\n", tabTitle, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 3, "Tabs not visited in this many days are considered old")
	cmd.Flags().Bool("debug", false, "Print tab/history debug info to stderr")
	return cmd
}
