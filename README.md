# Kitty-Meow

Meow is a kitty terminal extension for loading projects from a local directory, or even github.

Each kitty tab represents a project. Meow allows you to easily load and switch between projects.

If you've used tmux sessions. This is similar to switching between sessions, but allows you to
create new sesssions as well.

# Installation

```
git clone git@github.com:taylorzr/kitty-meow.git ~/.config/kitty/meow
pip install pyfzf
```

Depends on binaries [fzf](https://github.com/junegunn/fzf/), and [jq](https://github.com/stedolan/jq) as well.


# Configuration

You'll need to create bindings for loading projects, and switching between projects. You'll also
need to set your github token.

## new projects mapping

Create a binding for loading projects. The pattern is:

```
# ~/.config/kitty/kitty.conf
map <keys> kitten meow/load_project.py <directory where you put your repos> <github_org>
```

For example, I use:
```
# ~/.config/kitty/kitty.conf
map ctrl+n kitten meow/load_project.py $HOME/code/ my_cool_org
```


## github auth

You need to create a github token, and set it as env GITHUB_TOKEN. I do this by including a file
with secrets:

```
# ~/.config/kitty/kitty.conf
include ./dont_commit_me.conf
```

```
# ~/.config/kitty/dont_commit_me.conf
env GITHUB_TOKEN=<github_token>
```

You need to put env in your kitty config, not .zshrc. More about that here: https://sw.kovidgoyal.net/kitty/faq/#things-behave-differently-when-running-kitty-from-system-launcher-vs-from-another-terminal


## switching between projects

You probably also want a binding to switch between projects:
```
action_alias switch_project launch --type=overlay zsh -ic "kitty @ ls | jq -r '.[0].tabs | map(.title) | .[]' | fzf | xargs -I {} kitty @ focus-tab --match title:{}"
map ctrl+space switch_project
map ctrl+- goto_tab -1
```

# Usage

* ctrl+n
* fuzzy find any local project, and hit enter
* or then ctrl+g
* fuzzy find any github repo, and hit enter


# TODO

* record short video demo
* get_all_repos should get curent users repos
* get_all_repos should allow multiple orgs
* document and make configurable any mac/brew path assumptions
* caching for big orgs
  see what this dude did: https://mattorb.com/fuzzy-find-a-github-repository-part-deux/
