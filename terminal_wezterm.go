package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type WeztermTerminal struct{}

type weztermPane struct {
	WindowID int    `json:"window_id"`
	TabID    int    `json:"tab_id"`
	PaneID   int    `json:"pane_id"`
	Title    string `json:"title"`
}

func (w *WeztermTerminal) ListTabs() ([]Tab, error) {
	out, err := exec.Command("wezterm", "cli", "list", "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("wezterm cli list failed: %w", err)
	}

	var panes []weztermPane
	if err := json.Unmarshal(out, &panes); err != nil {
		return nil, fmt.Errorf("failed to parse wezterm output: %w", err)
	}

	// Deduplicate by tab_id, keeping first pane's title per tab
	seen := map[int]bool{}
	var tabs []Tab
	for _, p := range panes {
		if !seen[p.TabID] {
			seen[p.TabID] = true
			tabs = append(tabs, Tab{ID: p.TabID, Title: p.Title})
		}
	}
	return tabs, nil
}

func (w *WeztermTerminal) CloseTab(title string) error {
	tabs, err := w.ListTabs()
	if err != nil {
		return err
	}
	for _, t := range tabs {
		if t.Title == title {
			return exec.Command("wezterm", "cli", "close-tab", "--tab-id", strconv.Itoa(t.ID)).Run()
		}
	}
	return fmt.Errorf("tab %q not found", title)
}

func (w *WeztermTerminal) FocusTab(title string) error {
	tabs, err := w.ListTabs()
	if err != nil {
		return err
	}
	for _, t := range tabs {
		if t.Title == title {
			return exec.Command("wezterm", "cli", "activate-tab", "--tab-id", strconv.Itoa(t.ID)).Run()
		}
	}
	return fmt.Errorf("tab %q not found", title)
}

func (w *WeztermTerminal) sendText(paneID, text string) error {
	cmd := exec.Command("wezterm", "cli", "send-text", "--pane-id", paneID, "--no-paste")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (w *WeztermTerminal) NewTab(title, cwd string) error {
	out, err := exec.Command("wezterm", "cli", "spawn", "--cwd", cwd).Output()
	if err != nil {
		return err
	}
	firstPaneID := strings.TrimSpace(string(out))

	// Set tab title
	if err := exec.Command("wezterm", "cli", "set-tab-title", "--pane-id", firstPaneID, title).Run(); err != nil {
		return err
	}

	template := viper.GetStringSlice("template")
	if len(template) == 0 {
		return nil
	}

	if template[0] != "$SHELL" {
		if err := w.sendText(firstPaneID, template[0]+"\n"); err != nil {
			return err
		}
	}

	for _, entry := range template[1:] {
		out, err := exec.Command("wezterm", "cli", "split-pane", "--pane-id", firstPaneID, "--cwd", cwd).Output()
		if err != nil {
			return err
		}
		if entry != "$SHELL" {
			paneID := strings.TrimSpace(string(out))
			if err := w.sendText(paneID, entry+"\n"); err != nil {
				return err
			}
		}
	}

	// Return focus to the first pane
	if len(template) > 1 {
		if err := exec.Command("wezterm", "cli", "activate-pane", "--pane-id", firstPaneID).Run(); err != nil {
			return err
		}
	}

	return nil
}
