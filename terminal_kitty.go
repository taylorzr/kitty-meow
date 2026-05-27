package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

type KittyTerminal struct{}

type kittyWindow struct {
	ID int `json:"id"`
}

type kittyTab struct {
	ID      int           `json:"id"`
	Title   string        `json:"title"`
	Windows []kittyWindow `json:"windows"`
}

type kittyOSWindow struct {
	Tabs []kittyTab `json:"tabs"`
}

func (k *KittyTerminal) ListTabs() ([]Tab, error) {
	out, err := exec.Command("kitty", "@", "ls").Output()
	if err != nil {
		return nil, fmt.Errorf("kitty @ ls failed: %w", err)
	}

	var windows []kittyOSWindow
	if err := json.Unmarshal(out, &windows); err != nil {
		return nil, fmt.Errorf("failed to parse kitty output: %w", err)
	}

	var tabs []Tab
	for _, w := range windows {
		for _, t := range w.Tabs {
			tabs = append(tabs, Tab{ID: t.ID, Title: t.Title})
		}
	}
	return tabs, nil
}

// tabIDByTitle returns the ID of the tab with the given title, or -1 if not found.
func (k *KittyTerminal) tabIDByTitle(title string) int {
	tabs, err := k.ListTabs()
	if err != nil {
		return -1
	}
	for _, t := range tabs {
		if t.Title == title {
			return t.ID
		}
	}
	return -1
}

func (k *KittyTerminal) CloseTab(title string) error {
	if id := k.tabIDByTitle(title); id >= 0 {
		return exec.Command("kitty", "@", "close-tab", "--match", fmt.Sprintf("id:%d", id)).Run()
	}
	return fmt.Errorf("tab %q not found", title)
}

func (k *KittyTerminal) FocusTab(title string) error {
	if id := k.tabIDByTitle(title); id >= 0 {
		return exec.Command("kitty", "@", "focus-tab", "--match", fmt.Sprintf("id:%d", id)).Run()
	}
	return fmt.Errorf("tab %q not found", title)
}

func (k *KittyTerminal) sendText(windowID, command string) error {
	return exec.Command("kitty", "@", "send-text", "--match", "id:"+windowID, command+"\n").Run()
}

func (k *KittyTerminal) NewTab(title, cwd string) error {
	out, err := exec.Command("kitty", "@", "launch", "--type", "tab", "--title", title, "--tab-title", title, "--cwd", cwd).Output()
	if err != nil {
		return err
	}
	firstWindowID := strings.TrimSpace(string(out))

	template := viper.GetStringSlice("template")
	if len(template) == 0 {
		return nil
	}

	if template[0] != "$SHELL" {
		if err := k.sendText(firstWindowID, template[0]); err != nil {
			return err
		}
	}

	for _, entry := range template[1:] {
		out, err := exec.Command("kitty", "@", "launch", "--type", "window", "--dont-take-focus", "--cwd", cwd, "--title=current").Output()
		if err != nil {
			return err
		}
		if entry != "$SHELL" {
			windowID := strings.TrimSpace(string(out))
			if err := k.sendText(windowID, entry); err != nil {
				return err
			}
		}
	}

	return nil
}
