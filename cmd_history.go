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
	return &cobra.Command{
		Use:   "history",
		Short: "List recently visited projects from history",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := readHistory()
			if err != nil {
				return err
			}
			if len(items) > 0 {
				fmt.Print(strings.Join(items, "\n") + "\n")
			}
			return nil
		},
	}
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
		fields := strings.Fields(lines[i])
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		if seen[name] {
			continue
		}
		seen[name] = true
		alias := aliasFor(name)
		var entry string
		if len(fields) >= 2 {
			if t, err := parseTimestamp(fields[1]); err == nil {
				rel := dim(relativeTime(now.Sub(t)))
				entry = fmt.Sprintf("%s  %-25s %s", aliasCol(alias), name, rel)
			}
		}
		if entry == "" {
			entry = fmt.Sprintf("%s  %s", aliasCol(alias), name)
		}
		items = append(items, entry)
	}
	return items, nil
}
