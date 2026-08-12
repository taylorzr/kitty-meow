package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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

// parseGitURL extracts the repo name from an SSH (git@host:owner/repo(.git))
// or HTTPS (https://host/owner/repo(.git)) URL. Returns ok=false for anything
// that doesn't look like a git URL.
func parseGitURL(url string) (name string, ok bool) {
	u := strings.TrimSpace(url)
	var path string
	switch {
	case strings.HasPrefix(u, "git@"):
		// git@host:owner/repo(.git)
		parts := strings.SplitN(u, ":", 2)
		if len(parts) != 2 {
			return "", false
		}
		path = parts[1]
	case strings.HasPrefix(u, "https://"):
		// https://host/owner/repo(.git)
		rest := strings.TrimPrefix(u, "https://")
		if i := strings.Index(rest, "/"); i >= 0 {
			path = rest[i+1:]
		}
	default:
		return "", false
	}
	// Require at least owner/repo so user pages ("https://github.com/user")
	// aren't mistaken for repositories.
	if path == "" || !strings.Contains(path, "/") {
		return "", false
	}
	return gitRepoName(path)
}

// gitRepoName returns the last path segment of a repo path, stripped of a
// trailing ".git". A trailing slash is ignored.
func gitRepoName(path string) (string, bool) {
	path = strings.TrimSuffix(path, "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	if path == "" {
		return "", false
	}
	path = strings.TrimSuffix(path, ".git")
	if path == "" {
		return "", false
	}
	return path, true
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
