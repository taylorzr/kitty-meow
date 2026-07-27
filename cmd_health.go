package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check that required tools and auth are working",
		RunE: func(cmd *cobra.Command, args []string) error {
			ok := true

			// fzf
			fzf := resolveBin(viper.GetString("fzf"))
			if path, err := exec.LookPath(fzf); err == nil {
				fmt.Printf("  fzf:  %s\n", path)
			} else {
				fmt.Fprintf(os.Stderr, "  fzf:  \033[31mnot found\033[0m (looked in PATH and /opt/homebrew/bin, /usr/local/bin)\n")
				ok = false
			}

			// gh
			gh := resolveBin(viper.GetString("gh"))
			if path, err := exec.LookPath(gh); err == nil {
				fmt.Printf("  gh:   %s\n", path)
			} else {
				fmt.Fprintf(os.Stderr, "  gh:   \033[31mnot found\033[0m (looked in PATH and /opt/homebrew/bin, /usr/local/bin)\n")
				ok = false
			}

			// GitHub auth
			token, err := githubToken()
			if err != nil {
				fmt.Fprintf(os.Stderr, "  auth: \033[31m%s\033[0m\n", err)
				ok = false
			} else {
				fmt.Printf("  auth: ok (%s...)\n", token[:min(len(token), 8)])
			}

			// Terminal
			t := viper.GetString("terminal")
			if t == "" {
				t = os.Getenv("TERM_PROGRAM")
			}
			if t == "" {
				t = "kitty (default)"
			}
			fmt.Printf("  term: %s\n", t)

			if !ok {
				fmt.Fprintf(os.Stderr, "\nSome checks failed. See above for details.\n")
				os.Exit(1)
			}
			fmt.Println("\nAll checks passed.")
			return nil
		},
	}
}
