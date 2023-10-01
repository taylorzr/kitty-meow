# Kitty-Meow

Meow is a kitty terminal extension for working with projects, where each kitty tab is a different
project. It allows you to fuzzy switch between projects, and load them either from local directories or github.

If you've used tmux, this is similar to switching between sessions, but allows you to
create new sessions as well.

![Meow Screenshot](screenshot.png)

## Usage

Call your project mapping, e.g. ctrl-p, and hit enter to select.  Initially, tabs & local projects are listed, but you can show remote, local project only, or tabs only.

On select

* if the project is already in a tab, meow switches to that tab
* if the project is a local dir, meow creates a new tab
* if the project is github, meow clones to the first --dir, and creates a new tab


## Installation

```sh
git clone git@github.com:taylorzr/kitty-meow.git ~/.config/kitty/meow
```

Depends on [fzf](https://github.com/junegunn/fzf/) and [jq](https://github.com/stedolan/jq).

## Getting Started

You'll need to:

* create mappings
* set your github token as env

For example:

```conf
# ~/.config/kitty/kitty.conf

env GITHUB_TOKEN=<github_token>
env BIN_PATH=/opt/homebrew/bin/ # probably only needed on macs
map ctrl+p kitten meow/load_project.py --dir $HOME/code/ --user my_cool_self --org my_cool_org
map ctrl+shift+n kitten meow/new_project.py --dir $HOME/code/
map ctrl+shift+x kitten meow/kill_old_projects.py
map ctrl+shift+g kitten meow/cache_all_repos.py --org my_cool_org
map ctrl+- goto_tab -1
```

### kitty mappings

#### Loading projects

Create a mapping for loading projects. The pattern is:

```conf
# ~/.config/kitty/kitty.conf

map ctrl+p kitten meow/load_project.py --dir $HOME/code/ --user <you> --org <github_org>
```

--dir can be provided multiple times.

* when a dir ends in /, meow shows all it's subdirs
* otherwise, meow only shows that specific dir

For example, I use:

```conf
# ~/.config/kitty/kitty.conf

map ctrl+p kitten meow/load_project.py --dir $HOME/code/ --dir $HOME --dir $HOME/.config/kitty/meow --org my_cool_org
```

On mac, paths are goofy. You proabably need to set env BIN_PATH as well. This should be the dir
containing and fzf.

```conf
# ~/.config/kitty/kitty.conf

env BIN_PATH=/opt/homebrew/bin/
```

#### Caching github repositories

Something about big github orgs, talk about how you don't have to use caching. If file doesn't exist
it'll fetch from github everytime.

```conf
map ctrl+shift+g kitten meow/cache_all_repos.py --org my_cool_org
```

### github auth

You need to create a github token, and set it as env GITHUB_TOKEN. Because I commit kitty.conf to my
dotfiles, I put any secrets in an extra conf file:

```conf
# ~/.config/kitty/kitty.conf

include ./dont_commit_me.conf
```

```conf
# ~/.config/kitty/dont_commit_me.conf

env GITHUB_TOKEN=<github_token>
```

You need to put env in your kitty config, not .zshrc. More about that [here](https://sw.kovidgoyal.net/kitty/faq/#things-behave-differently-when-running-kitty-from-system-launcher-vs-from-another-terminal)

## Caching

Just like the load_project mapping, you can specify multiple users and orgs in your cache mapping. You might want these to be different than
users and orgs in your load_project mapping, because an org might have lots of repos, but your user
just a few.

Any uncached users/orgs repos will be loaded from github on every call to load_projects. And the cache never expires, you must call cache_all_repos to refresh it.

## TODO

* record short video demo
* configurable fzf bindings
* selectable dir to clone to?
  * some people might use 1 dir for work and one for personal?
* maybe use flags like --login=user=taylorzr --login=org=my_cool_org
* combine the scripts into one cli with subcommands
* configurable layout, shouldn't assume vim and 2 panes
