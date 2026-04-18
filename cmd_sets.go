package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type setConfig struct {
	Name     string   `mapstructure:"name"`
	Projects []string `mapstructure:"projects"`
}

func getSets() []setConfig {
	var sets []setConfig
	viper.UnmarshalKey("sets", &sets)
	return sets
}

func newSetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sets",
		Short: "List configured sets",
		RunE: func(cmd *cobra.Command, args []string) error {
			sets := getSets()
			if len(sets) == 0 {
				return nil
			}
			var items []string
			for _, b := range sets {
				items = append(items, fmt.Sprintf("%-30s \033[2m%s\033[0m", b.Name, strings.Join(b.Projects, ", ")))
			}
			fmt.Print(strings.Join(items, "\n") + "\n")
			return nil
		},
	}
}
