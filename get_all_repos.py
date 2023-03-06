import os
import sys

import github

# TODO: multi-org

org = sys.argv[1]
cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{org}"

# TODO: provide a way to refresh cache, but maybe just do it from command line
# cause if we do it here, you have to wait for the thing to finish
# if you select an item while still loading, the cache won't be written
try:
    with open(cache, "r") as file:
        print(file.read())
except FileNotFoundError:
    repos = github.get_all_repos(org)

    with open(cache, "w") as file:
        file.write("\n".join(repos))
