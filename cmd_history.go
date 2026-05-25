package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "history",
		Short: "List recently visited projects from history",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := readHistory()
			if err != nil {
				return err
			}
			if mode != "" {
				for i, item := range items {
					items[i] = mode + "\t" + item
				}
			}
			if len(items) > 0 {
				fmt.Print(strings.Join(items, "\n") + "\n")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "Prefix each item with this mode (open|close) for use with fzf --with-nth")
	return cmd
}

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Fall back to Python's datetime.isoformat() format (no timezone)
	return time.ParseInLocation("2006-01-02T15:04:05.999999999", s, time.Local)
}

func readHistory() ([]string, error) {
	data, err := os.ReadFile(filepath.Join(meowDir, "history"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	seen := map[string]bool{}
	now := time.Now()
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var items []string
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if line == "" {
			continue
		}
		var name, tsStr string
		if tab := strings.IndexByte(line, '\t'); tab != -1 {
			// New format: name\ttimestamp
			name = line[:tab]
			tsStr = line[tab+1:]
		} else {
			// Legacy format: name timestamp (single space, timestamp is last field)
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			if len(fields) >= 2 {
				name = strings.Join(fields[:len(fields)-1], " ")
				tsStr = fields[len(fields)-1]
			} else {
				name = fields[0]
			}
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		alias := aliasFor(name)
		var entry string
		if tsStr != "" {
			if t, err := parseTimestamp(tsStr); err == nil {
				rel := dim(relativeTime(now.Sub(t)))
				entry = aliasCol(alias) + "\t" + nameCol(name) + "\t" + rel
			}
		}
		if entry == "" {
			entry = aliasCol(alias) + "\t" + name
		}
		items = append(items, entry)
	}
	return items, nil
}
