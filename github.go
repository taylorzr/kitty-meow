package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func readCache(owner string) ([]repo, error) {
	data, err := os.ReadFile(cacheFile(owner))
	if err != nil {
		return nil, err
	}
	var repos []repo
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			repos = append(repos, repo{Name: parts[0], SSHUrl: parts[1]})
		}
	}
	return repos, nil
}

func writeCache(owner string, repos []repo) error {
	if err := os.MkdirAll(meowDir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	var lines []string
	for _, r := range repos {
		lines = append(lines, r.Name+" "+r.SSHUrl)
	}
	return os.WriteFile(cacheFile(owner), []byte(strings.Join(lines, "\n")), 0644)
}

func cachedListRepos(owner string, refresh bool) ([]repo, error) {
	if !refresh {
		if repos, err := readCache(owner); err == nil {
			logf("cachedListRepos: cache hit for %q (%d repos)", owner, len(repos))
			return repos, nil
		}
		logf("cachedListRepos: cache miss for %q, fetching from GitHub", owner)
	} else {
		logf("cachedListRepos: refresh requested for %q", owner)
	}
	repos, err := listRepos(owner)
	if err != nil {
		logError(fmt.Sprintf("cachedListRepos: listRepos %q", owner), err)
		return nil, err
	}
	// Always persist fetched repos; log but don't fail on write errors.
	logError(fmt.Sprintf("cachedListRepos: writeCache %q", owner), writeCache(owner, repos))
	return repos, nil
}

type repo struct {
	Name   string `json:"name"`
	SSHUrl string `json:"sshUrl"`
}

type pageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type repositories struct {
	Nodes    []repo   `json:"nodes"`
	PageInfo pageInfo `json:"pageInfo"`
}

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

const reposQuery = `
query($login: String!, $cursor: String) {
    repositoryOwner(login: $login) {
        repositories(first: 100, after: $cursor) {
            nodes { name sshUrl }
            pageInfo { endCursor hasNextPage }
        }
    }
}`

const viewerQuery = `
query {
    viewer {
        login
        organizations(first: 100) {
            nodes { login }
        }
    }
}`

type viewerResult struct {
	Login         string `json:"login"`
	Organizations struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"organizations"`
}

func fetchViewer() (viewerResult, error) {
	token, err := githubToken()
	if err != nil {
		return viewerResult{}, err
	}

	body, _ := json.Marshal(graphqlRequest{Query: viewerQuery})
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if err != nil {
		return viewerResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return viewerResult{}, err
	}
	data, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		logError("fetchViewer: close response body", closeErr)
	}
	if err != nil {
		return viewerResult{}, err
	}
	if resp.StatusCode != 200 {
		return viewerResult{}, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, data)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		return viewerResult{}, err
	}
	if errs, ok := result["errors"]; ok {
		return viewerResult{}, fmt.Errorf("GraphQL errors: %s", errs)
	}
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(result["data"], &outer); err != nil {
		return viewerResult{}, fmt.Errorf("unexpected GitHub API response shape (data): %w", err)
	}
	var viewer viewerResult
	if err := json.Unmarshal(outer["viewer"], &viewer); err != nil {
		return viewerResult{}, fmt.Errorf("unexpected GitHub API response shape (viewer): %w", err)
	}
	return viewer, nil
}

func githubToken() (string, error) {
	// Try gh CLI on PATH (works in normal terminals).
	if token, err := ghAuthToken("gh"); err == nil {
		return token, nil
	}
	// When launched from a keybinding (e.g. kitty), the shell may not be a login
	// shell so gh isn't on PATH. Try common install locations.
	for _, path := range []string{
		"/opt/homebrew/bin/gh", // macOS Apple Silicon
		"/usr/local/bin/gh",    // macOS Intel / Linux Homebrew
	} {
		if token, err := ghAuthToken(path); err == nil {
			return token, nil
		}
	}
	// Official gh CLI env vars.
	if token := os.Getenv("GH_TOKEN"); token != "" {
		logf("githubToken: using GH_TOKEN env var")
		return token, nil
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		logf("githubToken: using GITHUB_TOKEN env var")
		return token, nil
	}
	return "", fmt.Errorf("no GitHub token found: install gh and run `gh auth login`, or set GITHUB_TOKEN")
}

func ghAuthToken(cmd string) (string, error) {
	out, err := exec.Command(cmd, "auth", "token").Output()
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("empty token from %s", cmd)
	}
	logf("githubToken: using token from %s", cmd)
	return token, nil
}

func listRepos(owner string) ([]repo, error) {
	logf("listRepos: fetching repos for %q", owner)
	token, err := githubToken()
	if err != nil {
		return nil, err
	}

	var allRepos []repo
	var cursor *string

	for {
		vars := map[string]any{"login": owner, "cursor": cursor}
		body, _ := json.Marshal(graphqlRequest{Query: reposQuery, Variables: vars})

		req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			logError("listRepos: close response body", err)
		}
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, data)
		}

		var result map[string]json.RawMessage
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		if errs, ok := result["errors"]; ok {
			return nil, fmt.Errorf("GraphQL errors: %s", errs)
		}

		var outer map[string]json.RawMessage
		if err := json.Unmarshal(result["data"], &outer); err != nil {
			return nil, fmt.Errorf("unexpected GitHub API response shape (data): %w", err)
		}
		var repoOwner map[string]json.RawMessage
		if err := json.Unmarshal(outer["repositoryOwner"], &repoOwner); err != nil {
			return nil, fmt.Errorf("unexpected GitHub API response shape (repositoryOwner): %w", err)
		}
		var reposPage repositories
		if err := json.Unmarshal(repoOwner["repositories"], &reposPage); err != nil {
			return nil, fmt.Errorf("unexpected GitHub API response shape (repositories): %w", err)
		}

		allRepos = append(allRepos, reposPage.Nodes...)
		logf("listRepos: fetched page, total so far=%d hasNextPage=%v", len(allRepos), reposPage.PageInfo.HasNextPage)

		if !reposPage.PageInfo.HasNextPage {
			break
		}
		cursor = &reposPage.PageInfo.EndCursor
	}

	logf("listRepos: done, total=%d repos for %q", len(allRepos), owner)
	return allRepos, nil
}
