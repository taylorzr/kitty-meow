package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCacheCmd() *cobra.Command {
	var modeRemove bool
	var modeList bool

	cmd := &cobra.Command{
		Use:   "cache [owner]",
		Short: "Fetch and cache GitHub repos",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if modeList {
				entries, err := os.ReadDir(meowDir)
				if err != nil {
					return err
				}
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), "cache_") {
						info, err := e.Info()
						if err != nil {
							return err
						}
						owner := strings.TrimPrefix(e.Name(), "cache_")
						fmt.Printf("%-30s %s\n", owner, relativeTime(time.Since(info.ModTime())))
					}
				}
				return nil
			}

			if len(args) == 0 {
				for _, r := range viper.GetStringSlice("github") {
					if _, err := os.Stat(cacheFile(r)); os.IsNotExist(err) {
						fmt.Fprintf(os.Stderr, "no cache for %s, skipping (run 'meow cache %s' to create)\n", r, r)
						continue
					}
					fmt.Fprintf(os.Stderr, "fetching repos for %s...\n", r)
					repos, err := listRepos(r)
					if err != nil {
						return err
					}
					if err := writeCache(r, repos); err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "cached %d repos for %s\n", len(repos), r)
				}
				return nil
			}

			owner := args[0]

			if modeRemove {
				if len(args) == 0 {
					return fmt.Errorf("--remove requires an owner argument")
				}

				if err := os.Remove(cacheFile(owner)); err != nil {
					return fmt.Errorf("could not remove cache: %w", err)
				}
				fmt.Fprintf(os.Stderr, "removed cache for %s\n", owner)
				return nil
			}

			fmt.Fprintf(os.Stderr, "fetching repos for %s...\n", owner)
			repos, err := listRepos(owner)
			if err != nil {
				return err
			}
			if err := writeCache(owner, repos); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "cached %d repos for %s\n", len(repos), owner)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&modeRemove, "remove", "r", false, "Remove the cache for the given owner")
	cmd.Flags().BoolVarP(&modeList, "list", "l", false, "Show when each cache was last fetched")
	return cmd
}
