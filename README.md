# meow

# Installation

```
git clone git@github.com:taylorzr/meow.git ~/.config/kitty/meow
```

Also requires fzf, and jq.


# Configuration

## new projects mapping

Create a binding for loading new_projects. The pattern is:

```
# ~/.config/kitty/kitty/conf
map <keys> kitten meow/new_project.py <directory where you put your repos> <github_org> 
```

For example, I use:
```
# ~/.config/kitty/kitty/conf
map ctrl+n kitten meow/new_project.py /Users/taylorzr/code/ <github_org>
```

## github auth

You need to create a github token, and set it as env GITHUB_TOKEN. I do this by including a file
with my secrets:

```
# ~/.config/kitty/kitty.conf
include ./dont_commit_me.conf
```

```
# ~/.config/kitty/dont_commit_me.conf
env GITHUB_TOKEN=<github_token>
```


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

