package main

import "testing"

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		url      string
		wantName string
		wantOK   bool
	}{
		// SSH
		{"git@github.com:rigwild/mcp-server-amazon.git", "mcp-server-amazon", true},
		{"git@github.com:owner/repo.git", "repo", true},
		{"git@github.com:owner/repo", "repo", true},
		{"git@gitlab.com:group/subgroup/repo.git", "repo", true},
		// HTTPS
		{"https://github.com/rigwild/mcp-server-amazon.git", "mcp-server-amazon", true},
		{"https://github.com/owner/repo.git", "repo", true},
		{"https://github.com/owner/repo", "repo", true},
		{"https://github.com/owner/repo/", "repo", true},
		{"https://gitlab.com/group/subgroup/repo.git", "repo", true},
		// Invalid
		{"", "", false},
		{"hello", "", false},
		{"https://example.com", "", false},
		{"http://github.com/owner/repo.git", "", false}, // http:// not supported
		{"git@github.com", "", false},                   // no owner/repo path
		{"git@github.com:", "", false},                  // empty path
		{"https://github.com/owner/", "owner", true},    // trailing slash ignored
	}

	for _, tt := range tests {
		name, ok := parseGitURL(tt.url)
		if ok != tt.wantOK || name != tt.wantName {
			t.Errorf("parseGitURL(%q) = (%q, %v), want (%q, %v)", tt.url, name, ok, tt.wantName, tt.wantOK)
		}
	}
}
