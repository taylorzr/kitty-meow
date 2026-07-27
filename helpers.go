package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// commonBinPaths returns common install locations for a binary on macOS and Linux.
func commonBinPaths(name string) []string {
	return []string{
		filepath.Join("/opt/homebrew/bin", name), // macOS Apple Silicon
		filepath.Join("/usr/local/bin", name),    // macOS Intel / Linux Homebrew
	}
}

// resolveBin finds an executable by trying the given name, then common install
// locations. Returns the resolved path or the original name if not found.
func resolveBin(name string) string {
	if _, err := exec.LookPath(name); err == nil {
		return name
	}
	for _, p := range commonBinPaths(name) {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return name
}

func relativeTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes())
	switch {
	case minutes < 1:
		return "just now"
	case minutes < 60:
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case hours < 48:
		return "1 day ago"
	case hours < 24*7:
		return fmt.Sprintf("%d days ago", hours/24)
	case hours < 24*30:
		weeks := hours / (24 * 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case hours < 24*365:
		months := hours / (24 * 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := hours / (24 * 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
