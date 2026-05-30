package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/spf13/viper"
)


func recordHistory(name string) error {
	if err := os.MkdirAll(meowDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(meowDir, "history")

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	lines = append(lines, name+"\t"+time.Now().Format(time.RFC3339))

	if len(lines) > viper.GetInt("history_size") {
		lines = lines[len(lines)-viper.GetInt("history_size"):]
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// readLastViews returns a map of project name → last visited time from history.
func readLastViews() map[string]time.Time {
	views := map[string]time.Time{}
	data, err := os.ReadFile(filepath.Join(meowDir, "history"))
	if err != nil {
		return views
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		name, timestamp := parseHistoryLine(line)
		if name != "" && timestamp != "" {
			if t, err := parseTimestamp(timestamp); err == nil {
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
		var name, timestamp string
		name, timestamp = parseHistoryLine(line)
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		alias := aliasFor(name)
		var entry string
		if timestamp != "" {
			if t, err := parseTimestamp(timestamp); err == nil {
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

// parseHistoryLine parses a history file line into (name, timestamp).
// Supports both new tab-delimited format ("name\ttimestamp") and
// legacy space-separated format ("name timestamp").
func parseHistoryLine(line string) (string, string) {
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

func parseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Fall back to Python's datetime.isoformat() format (no timezone)
	return time.ParseInLocation("2006-01-02T15:04:05.999999999", s, time.Local)
}
