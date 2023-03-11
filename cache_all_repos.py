import github
import argparse
import os
from typing import List

# TODO: can do relative import by file path?
# Need this so we can relative import github
# sys.path.append("~/.config/kitty/meow")


parser = argparse.ArgumentParser(description="meow")

parser.add_argument(
    "--org",
    dest="orgs",
    action="append",
    default=[],
    help="look for repos in these github orgs",
)

parser.add_argument(
    "--user",
    dest="users",
    action="append",
    default=[],
    help="look for repos for these github users",
)


def main(args: List[str]) -> str:
    opts = parser.parse_args(args[1:])

    for user in opts.users:
        cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{user}"
        repos = github.get_repos(user, type="user")
        with open(cache, "w") as file:
            file.write("\n".join(repos))

    for org in opts.orgs:
        cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{org}"
        repos = github.get_repos(org, type="organization")
        with open(cache, "w") as file:
            file.write("\n".join(repos))
