# Kitty-Meow

Meow is a kitty terminal tool for working with projects, where each kitty tab is a different
project. It allows you to fuzzy switch between projects, and load them either from local directories or github.

If you've used tmux, this is similar to switching between sessions, but allows you to
easily create new sessions as well.

![Meow Screenshot](screenshot.png)

## Basic Usage

[1-minute demo](https://www.youtube.com/watch?v=Qm8Xl4GAylI)

Call your switch mapping, e.g. ctrl-space, type to filter, and hit enter to select. Initially, tabs
& local projects are listed, but you can show only remote, or local, or open.


On select

- if the project is already in a tab, meow switches to that tab
- if the project is a local dir, meow creates a new tab
- if the project is github, meow clones to the first --dir, and creates a new tab


## Getting Started

### Installation

Depends on fzf see [fzf installation](https://github.com/junegunn/fzf/#installation).

**Download the latest release binary** (Linux/macOS):

```sh
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -s https://api.github.com/repos/taylorzr/kitty-meow/releases/latest \
  | grep browser_download_url \
  | grep "${OS}_${ARCH}" \
  | cut -d'"' -f4 \
  | xargs curl -sSL \
  | tar -xz -C ~/.local/bin
```

Or install with Go:

```sh
go install github.com/taylorzr/kitty-meow@latest
```

Ensure your install dir is in your path (`~/.local/bin` or `~/go/bin`).


### Configuration


1. configure `~/.config/kitty-meow/meow.toml`
2. update `~/.config/kitty/kitty.conf`
    - create keyboard shortcut for switching
    - optionally set `GITHUB_TOKEN` env (not needed if using `gh` CLI)

### meow.toml

Set at least dirs and github. More options can be seen in meow.example.toml.

```toml
# ~/.config/kitty-meow/meow.toml

dirs  = [
  "~/code/", # dirs ending in / list all projects within
  "~/.config/kitty-meow",  # otherwise the dir is treated as a single project
]

github = ["taylorzr", "AquaTeenHungerForce"]
```

If no dirs are set, projects will be listed and cloned to your home dir.

### kitty.conf

```conf
# ~/.config/kitty/kitty.conf

map ctrl+space kitty-meow switch
map ctrl+- goto_tab -1

# optional: this keeps tab title set to project name
window_title_format {tab.title}
# ensure you new pane binding includes `--title=current`
map ctrl+enter launch --cwd=current --title=current
```

## Caching github repositories

If you work in an org with lots of repos, listing remote projects can be very slow. So you can
cache the list with `kitty-meow cache <owner>` e.g. `kitty-meow cache taylorzr`.

- list the current caches with `kitty-meow cache --list`
- refresh all the caches with `kitty-meow cache`
- refresh a specific with `kitty-meow cache <owner>`
- delete a cache with `kitty-meow cache --remove <owner>`

Caches aren't automatically updated, so re-run `kitty-meow cache` as needed.

## Github Auth

kitty-meow will use `gh auth token` automatically if the [gh CLI](https://cli.github.com/) is
installed and authenticated. This is the recommended approach:

```sh
gh auth login
```

Alternatively, set `GITHUB_TOKEN` as an env var. Note it must be set in your kitty config, not
`.zshrc`. More about that
[here](https://sw.kovidgoyal.net/kitty/faq/#things-behave-differently-when-running-kitty-from-system-launcher-vs-from-another-terminal).

Because I commit kitty.conf to my dotfiles, I put any secrets in an extra conf file:

```conf
# ~/.config/kitty/kitty.conf

include ./dont_commit_me.conf
```

```conf
# ~/.config/kitty/dont_commit_me.conf

env GITHUB_TOKEN=<github_token>
```

## Migrating from the Python version

The Go rewrite replaces the collection of Python kittens (`projects.py`, `cache.py`, `kill.py`) with
a single `kitty-meow` binary. Configuration moves from flags on your kitty.conf keybindings into
`~/.config/kitty-meow/meow.toml`.

### 1. Install the binary

Follow the [Installation](#installation) steps above to install `kitty-meow`.

### 2. Create meow.toml

Create `~/.config/kitty-meow/meow.toml` and translate your old flags:

| Old kitty.conf flag | New meow.toml key |
|---|---|
| `--dir $HOME/code/` (repeatable) | `dirs = ["~/code/"]` |
| `--user taylorzr` | `github = ["taylorzr"]` |
| `--org AquaTeenHungerForce` | `github = ["AquaTeenHungerForce"]` |
| `env BIN_PATH=/opt/homebrew/bin/` | `fzf = "/opt/homebrew/bin/fzf"` |

Example — if your old kitty.conf had:

```conf
env BIN_PATH=/opt/homebrew/bin/
map ctrl+space kitten meow/projects.py load --dir $HOME/code/ --dir $HOME --user taylorzr --org AquaTeenHungerForce
```

Your new `~/.config/kitty-meow/meow.toml` would be:

```toml
dirs   = ["~/code/", "~/"]
github = ["taylorzr", "AquaTeenHungerForce"]
fzf    = "/opt/homebrew/bin/fzf"
```

### 3. Update kitty.conf keybindings

Replace all the old kitten mappings with a single `kitty-meow switch` mapping:

```conf
# Remove these old lines:
map ctrl+space   kitten meow/projects.py load --dir $HOME/code/ --user taylorzr
map ctrl+shift+n kitten meow/projects.py new --dir $HOME/code/
map ctrl+shift+g kitten meow/cache.py --org AquaTeenHungerForce
map ctrl+shift+x kitten meow/kill.py

# Replace with:
map ctrl+space kitty-meow switch
```

Project closing (`kill.py`) is now a built-in fzf binding inside `kitty-meow switch` — no
separate keybinding needed.

### 4. GitHub auth

The binary will automatically pick up a token from `gh auth token` if you have the
[gh CLI](https://cli.github.com/) installed. If not, `GITHUB_TOKEN` (or `GH_TOKEN`) env vars
still work. You can remove the `env GITHUB_TOKEN=...` line from kitty.conf if you use `gh`.

### 5. Caching

The cache command is now part of the binary:

```conf
# Old:
map ctrl+shift+g kitten meow/cache.py --org AquaTeenHungerForce

# New (optional — or just run kitty-meow cache from a terminal):
map ctrl+shift+g kitty-meow cache
```

### 6. Clean up

You can delete the old Python kitten directory:

```sh
rm -rf ~/.config/kitty/meow
```

## Experimental Wezterm Support

Added support for wezterm in addition to kitty. This is very much experimental, and may be removed in the future.

```lua
# ~/.config/wezterm/wezterm.lua

local wezterm = require 'wezterm'
local config = wezterm.config_builder()

config.default_prog = { 'zsh' }

config.keys = {
  {
    key = 'Space',
    mods = 'CTRL',
    action = wezterm.action.SplitPane {
      direction = 'Down',
      command = { args = { os.getenv("HOME") .. '/go/bin/kitty-meow', 'switch' } },
    },
  }
}

return config
```

## TODO

- testing in CI, github actions?
- tests
- fix project ordering when swapping explicitly to default
- support other git sources, like gitlab
- cmds to go out/in of projects like vim ctrl-o/i, maybe ctrl-shift-o/i?
