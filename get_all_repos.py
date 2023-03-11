import os
import argparse

import github

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


def get_repos(login, type):
    cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{login}"
    try:
        with open(cache, "r") as file:
            print(file.read())
    except FileNotFoundError:
        github.get_repos(login, type)


if __name__ == "__main__":
    opts = parser.parse_args()
    for user in opts.users:
        get_repos(login=user, type="user")
    for org in opts.orgs:
        get_repos(login=org, type="organization")
