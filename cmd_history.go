package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Fall back to Python's datetime.isoformat() format (no timezone)
	return time.ParseInLocation("2006-01-02T15:04:05.999999999", s, time.Local)
}

// parseHistoryLine parses a history file line into (name, tsStr).
// Supports both new tab-delimited format ("name\ttimestamp") and
// legacy space-separated format ("name timestamp").
func parseHistoryLine(line string) (name, tsStr string) {
	if tab := strings.IndexByte(line, '\t'); tab != -1 {
		return line[:tab], line[tab+1:]
	}
	fields := strings.Fields(line)
	switch len(fields) {
	case 0:
		return "", ""
	case 1:
		return fields[0], ""
	default:
		return strings.Join(fields[:len(fields)-1], " "), fields[len(fields)-1]
	}
}

// readLastViews returns a map of project name → last visited time from history.
func readLastViews() map[string]time.Time {
	views := map[string]time.Time{}
	data, err := os.ReadFile(filepath.Join(meowDir, "history"))
	if err != nil {
		return views
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		name, tsStr := parseHistoryLine(line)
		if name != "" && tsStr != "" {
			if t, err := parseTimestamp(tsStr); err == nil {
				views[name] = t
			}
		}
	}
	return views
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
		name, tsStr = parseHistoryLine(line)
		if name == "" {
			continue
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
