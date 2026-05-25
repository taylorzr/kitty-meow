package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var meowDir = expandHome("~/.config/kitty/meow")

func cacheFile(owner string) string {
	return filepath.Join(meowDir, "cache_"+owner)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func collapseHome(path string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(path, home, "~", 1)
}

type bind struct {
	key     string
	prompt  string
	command string
}

func bindsAndHeader(emoji string, binds []bind) (string, string) {
	var bindParts, headerParts []string
	for _, b := range binds {
		if b.key == "" {
			// FIXME: at least log an error or something
			continue
		}
		headerParts = append(headerParts, fmt.Sprintf("%s: %s", b.key, b.prompt))
		bindParts = append(bindParts, fmt.Sprintf("%s:change-prompt(%s %s > )+reload(%s)", b.key, emoji, b.prompt, b.command))
	}
	return strings.Join(bindParts, ","), strings.Join(headerParts, " | ")
}

// aliasFor returns the alias key for the given real project name, or "".
func aliasFor(name string) string {
	for alias, real := range viper.GetStringMapString("aliases") {
		if real == name {
			return alias
		}
	}
	return ""
}

// nameFor returns the real project name for the given alias, or the alias itself if not found.
func nameFor(alias string) string {
	aliases := viper.GetStringMapString("aliases")
	if real, ok := aliases[alias]; ok {
		return real
	}
	return alias
}

func main() {
	home, _ := os.UserHomeDir()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(meowDir)
	viper.SetDefault("dirs", []string{home + "/"})
	viper.SetDefault("template", []string{"$EDITOR", "$SHELL"})
	viper.SetDefault("fzf", "fzf")
	viper.SetDefault("history_size", 10000)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
		}
	}

	initTerminal()

	rootCmd := &cobra.Command{
		Use:   "meow",
		Short: "Kitty terminal project manager",
	}
	rootCmd.AddCommand(newProjectsCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newSetsCmd())
	rootCmd.AddCommand(newCacheCmd())
	rootCmd.AddCommand(newSwitchCmd())
	rootCmd.AddCommand(newCloseCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
