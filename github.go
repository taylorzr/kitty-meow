package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	var lines []string
	for _, r := range repos {
		lines = append(lines, r.Name+" "+r.SSHUrl)
	}
	return os.WriteFile(cacheFile(owner), []byte(strings.Join(lines, "\n")), 0644)
}

func cachedListRepos(owner string, refresh bool) ([]repo, error) {
	if !refresh {
		if repos, err := readCache(owner); err == nil {
			return repos, nil
		}
	}
	repos, err := listRepos(owner)
	if err != nil {
		return nil, err
	}
	if refresh {
		err = writeCache(owner, repos)
	}
	return repos, err
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

func listRepos(owner string) ([]repo, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
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

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
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
		json.Unmarshal(result["data"], &outer)
		var owner map[string]json.RawMessage
		json.Unmarshal(outer["repositoryOwner"], &owner)
		var reposPage repositories
		json.Unmarshal(owner["repositories"], &reposPage)

		allRepos = append(allRepos, reposPage.Nodes...)

		if !reposPage.PageInfo.HasNextPage {
			break
		}
		cursor = &reposPage.PageInfo.EndCursor
	}

	return allRepos, nil
}
