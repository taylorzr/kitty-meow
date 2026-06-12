package main

import (
	"os"

	"github.com/spf13/viper"
)

type Tab struct {
	ID    int
	Title string
}

type Terminal interface {
	ListTabs() ([]Tab, error)
	FocusTab(title string) error
	NewTab(title, cwd string, template []string) error
	CloseTab(title string) error
}

var term Terminal

func initTerminal() {
	t := viper.GetString("terminal")
	if t == "" {
		t = os.Getenv("TERM_PROGRAM")
	}
	switch t {
	case "WezTerm":
		term = &WeztermTerminal{}
	default:
		term = &KittyTerminal{}
	}
}
