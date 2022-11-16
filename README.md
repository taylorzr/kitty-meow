# Kitty-Meow

Meow is a kitty terminal extension for working with projects. It allows you to easily switch between
projects, and load them either from local directories or github.

If you've used tmux sessions. This is similar to switching between sessions, but allows you to
create new sessions as well.

# Installation

```
git clone git@github.com:taylorzr/kitty-meow.git ~/.config/kitty/meow
pip install pyfzf
```

Depends on [fzf](https://github.com/junegunn/fzf/) and [jq](https://github.com/stedolan/jq).


# Configuration

You'll need to:
* create mapping for loading projects
* create mapping for switching between projects
* set your github token as env

## load projects mapping

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

On mac, paths are goofy. You proabably need to set env BIN_PATH as well:

```
# ~/.config/kitty/kitty.conf
env BIN_PATH=/opt/homebrew/bin/
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

You need to put env in your kitty config, not .zshrc. More about that [here](https://sw.kovidgoyal.net/kitty/faq/#things-behave-differently-when-running-kitty-from-system-launcher-vs-from-another-terminal)


## switching between projects

You probably also want a binding to switch between projects:
```
map ctrl+space launch --type=overlay zsh -ic "kitty @ ls | jq -r '.[0].tabs | map(.title) | .[]' | fzf | xargs -I {} kitty @ focus-tab --match title:{}"
map ctrl+- goto_tab -1
```

## all together

```
env GITHUB_TOKEN=<github_token>
env BIN_PATH=/opt/homebrew/bin/
map ctrl+space launch --type=overlay zsh -ic "kitty @ ls | jq -r '.[0].tabs | map(.title) | .[]' | fzf | xargs -I {} kitty @ focus-tab --match title:{}"
map ctrl+- goto_tab -1
map ctrl+n kitten meow/load_project.py $HOME/code/ my_cool_org
map ctrl+shift+t launch --type=overlay zsh -c "print 'tab title:' && read title && kitty @ set-tab-title $title"
```

# Usage

* ctrl+n
* fuzzy find any local project, and hit enter
* or then ctrl+g
* fuzzy find any github repo, and hit enter


# TODO

* record short video demo
* get_all_repos should get curent users repos
* get_all_repos should allow multiple orgs, or maybe even none
  * will probably need some cli flags, good python repo for this?
* document and make configurable any mac/brew path assumptions
* caching for big orgs
  see what this dude did: https://mattorb.com/fuzzy-find-a-github-repository-part-deux/
* something for renaming tabs?
* configurable binding for fetching github stuff (ctrl-g)
* load projects from other dirs, e.g. meow itself i keep in ~/.config/kitty/meow, not ~/code/meow
  * but then where would cloned projects go? guessing the first dir in user config?
