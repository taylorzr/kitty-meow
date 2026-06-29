package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configError string

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
			logf("WARNING bindsAndHeader: bind with empty key skipped (prompt: %q)", b.prompt)
			continue
		}
		headerParts = append(headerParts, fmt.Sprintf("%s: %s", b.key, b.prompt))
		bindParts = append(bindParts, fmt.Sprintf("%s:change-prompt(%s %s > )+reload(%s)", b.key, emoji, b.prompt, b.command))
	}
	return strings.Join(bindParts, ","), strings.Join(headerParts, " | ")
}

type projectConfig struct {
	Name     string   `mapstructure:"name"`
	Alias    string   `mapstructure:"alias"`
	Template []string `mapstructure:"template"`
}

func getProjects() []projectConfig {
	var projects []projectConfig
	viper.UnmarshalKey("projects", &projects)
	return projects
}

// aliasFor returns the alias for the given real project name, or "".
func aliasFor(name string) string {
	for _, p := range getProjects() {
		if p.Name == name && p.Alias != "" {
			return p.Alias
		}
	}
	return ""
}

// nameFor returns the real project name for the given alias, or the alias itself if not found.
func nameFor(alias string) string {
	for _, p := range getProjects() {
		if p.Alias == alias {
			return p.Name
		}
	}
	return alias
}

// templateFor returns the pane template for the given project name or alias.
// Falls back to the global template if no per-project template is set.
func templateFor(name string) []string {
	for _, p := range getProjects() {
		if p.Name == name || p.Alias == name {
			if len(p.Template) > 0 {
				return p.Template
			}
		}
	}
	return viper.GetStringSlice("template")
}

func main() {
	initLogger()

	home, _ := os.UserHomeDir()
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(meowDir)
	viper.SetDefault("dirs", []string{home + "/"})
	viper.SetDefault("template", []string{"$EDITOR", "$SHELL"})
	viper.SetDefault("fzf", "fzf")
	viper.SetDefault("history_size", 10000)
	viper.SetDefault("spacing.alias", 5)
	viper.SetDefault("spacing.name", 32)
	viper.SetDefault("hide_dot_dirs", true)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			configError = fmt.Sprintf("config error: %v", err)
		}
	}

	initTerminal()

	rootCmd := &cobra.Command{
		Use:   "meow",
		Short: "Kitty terminal project manager",
	}
	rootCmd.AddCommand(newProjectsCmd())
	rootCmd.AddCommand(newCacheCmd())
	rootCmd.AddCommand(newSwitchCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
