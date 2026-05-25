package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

			lastViewMap := readLastViews()
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			cutoff := today.AddDate(0, 0, -days)

			if debug {
				fmt.Fprintf(os.Stderr, "cutoff: %s\n", cutoff.Format(time.RFC3339))
				fmt.Fprintf(os.Stderr, "open tabs (%d):\n", len(tabs))
				for _, t := range tabs {
					last, seen := lastViewMap[nameFor(t.Title)]
					if seen {
						fmt.Fprintf(os.Stderr, "  %q last=%s old=%v\n", t.Title, last.Format(time.RFC3339), last.Before(cutoff))
					} else {
						fmt.Fprintf(os.Stderr, "  %q (not in history → old)\n", t.Title)
					}
				}
				fmt.Fprintf(os.Stderr, "lastViewMap (%d entries):\n", len(lastViewMap))
				for k, v := range lastViewMap {
					fmt.Fprintf(os.Stderr, "  %q → %s\n", k, v.Format(time.RFC3339))
				}
			}

			// Entry format: "tabTitle\taliasCol\tnameCol(displayName)\tdim(relTime)"
			// Field 1 (tabTitle) is hidden via --with-nth=2..; used for closing.
			var allTabs, oldTabs []string
			for _, t := range tabs {
				if t.Title == "" {
					continue
				}
				displayName := nameFor(t.Title)
				var alias string
				if displayName != t.Title {
					alias = t.Title // tab title is the alias
				} else {
					alias = aliasFor(t.Title)
				}
				realName := displayName

				var entry string
				if last, seen := lastViewMap[realName]; seen {
					rel := dim(relativeTime(now.Sub(last)))
					entry = t.Title + "\t" + aliasCol(alias) + "\t" + nameCol(displayName) + "\t" + rel
				} else {
					entry = t.Title + "\t" + aliasCol(alias) + "\t" + displayName
				}

				allTabs = append(allTabs, entry)
				last, seen := lastViewMap[realName]
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
				// Entry is "tabTitle\taliasCol\t..." — field 0 is the actual tab title.
				parts := strings.SplitN(ansiRe.ReplaceAllString(title, ""), "\t", 2)
				tabTitle := strings.TrimSpace(parts[0])
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
