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
	var mode string
	cmd := &cobra.Command{
		Use:   "sets",
		Short: "List configured sets",
		RunE: func(cmd *cobra.Command, args []string) error {
			sets := getSets()
			if len(sets) == 0 {
				return nil
			}
			var items []string
			prefix := ""
			if mode != "" {
				prefix = mode + "\t"
			}
			for _, b := range sets {
				items = append(items, prefix+aliasCol("")+"\t"+nameCol(b.Name)+"\t"+dim(strings.Join(b.Projects, ", ")))
			}
			fmt.Print(strings.Join(items, "\n") + "\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "Prefix each item with this mode (open|close) for use with fzf --with-nth")
	return cmd
}
