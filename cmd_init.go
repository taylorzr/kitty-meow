package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed meow.example.toml
var exampleConfig string

// parsedExample holds the comment headers for dirs/github (to print above the
// filled-in values) and the remaining hints (everything else, active lines
// commented out).
type parsedExample struct {
	dirsComments   string
	githubComments string
	hints          string
}

func parseExample() parsedExample {
	lines := strings.Split(exampleConfig, "\n")

	var dirsComments, githubComments []string
	var hints strings.Builder

	var commentBuf []string // comments buffered while waiting to see what follows
	inDirsBlock := false

	flush := func(commented bool) {
		for _, c := range commentBuf {
			if commented {
				hints.WriteString("# " + c + "\n")
			} else {
				hints.WriteString(c + "\n")
			}
		}
		commentBuf = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inDirsBlock {
			if trimmed == "]" {
				inDirsBlock = false
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			commentBuf = append(commentBuf, line)
			continue
		}

		// Active line — decide what to do based on which key it is.
		if strings.HasPrefix(trimmed, "dirs") {
			dirsComments = commentBuf
			commentBuf = nil
			inDirsBlock = true
			continue
		}

		if strings.HasPrefix(trimmed, "github") {
			githubComments = commentBuf
			commentBuf = nil
			continue // skip the active github line; init writes it
		}

		// Any other active line: flush buffered comments as-is, write line commented.
		flush(false)
		hints.WriteString("# " + line + "\n")
	}
	flush(false)

	join := func(ls []string) string {
		if len(ls) == 0 {
			return ""
		}
		return strings.Join(ls, "\n") + "\n"
	}

	return parsedExample{
		dirsComments:   join(dirsComments),
		githubComments: join(githubComments),
		hints:          strings.TrimRight(hints.String(), "\n") + "\n",
	}
}

func newInitCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactively create ~/.config/kitty-meow/meow.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(meowDir, "meow.toml")

			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", configPath)
			}
			if _, err := os.Stat(configPath); err == nil && force {
				backupPath := configPath + ".bak"
				if err := os.Rename(configPath, backupPath); err != nil {
					return fmt.Errorf("backing up existing config: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Backed up existing config to %s\n", backupPath)
			}

			scanner := bufio.NewScanner(os.Stdin)

			// --- Dirs ---
			defaultDirs := "~/code/"
			fmt.Printf("What dir(s) do you keep your projects? (space-separated, trailing / lists all subdirs) [%s]: ", defaultDirs)
			scanner.Scan()
			dirsInput := strings.TrimSpace(scanner.Text())
			if dirsInput == "" {
				dirsInput = defaultDirs
			}
			dirs := strings.Fields(dirsInput)

			// --- GitHub owners ---
			var github []string
			fmt.Fprintln(os.Stderr, "\nFetching your GitHub orgs...")
			viewer, err := fetchViewer()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not fetch GitHub orgs (%v)\nSkipping — you can add github = [...] to meow.toml manually.\n", err)
			} else {
				candidates := []string{viewer.Login}
				for _, org := range viewer.Organizations.Nodes {
					candidates = append(candidates, org.Login)
				}

				selected, err := fzfSelect(candidates)
				if err != nil || selected == "" {
					fmt.Fprintln(os.Stderr, "No GitHub owners selected.")
				} else {
					for _, line := range strings.Split(strings.TrimSpace(selected), "\n") {
						if line != "" {
							github = append(github, line)
						}
					}
				}
			}

			// --- Write config ---
			if err := os.MkdirAll(meowDir, 0755); err != nil {
				return fmt.Errorf("creating config dir: %w", err)
			}

			ex := parseExample()
			var buf bytes.Buffer

			buf.WriteString(ex.dirsComments)
			buf.WriteString("dirs = [")
			for i, d := range dirs {
				if i > 0 {
					buf.WriteString(", ")
				}
				fmt.Fprintf(&buf, "%q", d)
			}
			buf.WriteString("]\n")

			buf.WriteString("\n")
			buf.WriteString(ex.githubComments)
			if len(github) > 0 {
				buf.WriteString("github = [")
				for i, g := range github {
					if i > 0 {
						buf.WriteString(", ")
					}
					fmt.Fprintf(&buf, "%q", g)
				}
				buf.WriteString("]\n")
			} else {
				buf.WriteString("# github = []\n")
			}

			buf.WriteString("\n")
			buf.WriteString(ex.hints)

			if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Fprintf(os.Stderr, "\nWrote %s\n", configPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config")
	return cmd
}

func fzfSelect(items []string) (string, error) {
	fzfBin := viper.GetString("fzf")
	p := exec.Command(fzfBin, "--multi", "--reverse", "--prompt=🐈 meow > ", "--header=Select GitHub users/orgs to include (tab to select multiple)")
	p.Stdin = strings.NewReader(strings.Join(items, "\n"))
	p.Stderr = os.Stderr
	out, err := p.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
