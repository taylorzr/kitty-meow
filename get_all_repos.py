import os
import argparse
import subprocess
from urllib.parse import urlparse

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

parser.add_argument(
    "--ssh",
    dest="ssh",
    action="append",
    default=[],
    help="look for projects in these ssh paths",
)


def get_repos(login, type):
    cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{login}"
    try:
        with open(cache, "r") as file:
            print(file.read())
    except FileNotFoundError:
        github.get_repos(login, type)


def get_ssh_repos(url):
    # TODO: cache by host / path
    # cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{login}"
    uri = urlparse(url)
    dest = ""

    if uri.username:
        dest += uri.username
        dest += "@"

    dest += uri.hostname

    if uri.port:
        dest += f" -p {uri.port}"

    # TODO: error is we don't have the parts we need, at least a dest & path

    result = subprocess.run(f"ssh {dest} 'ls -d {uri.path.lstrip("/")}*'", shell=True, capture_output=True, text=True)

    if result.returncode != 0:
        print("Error:", result.stderr)

    for line in result.stdout.splitlines():
        if line:
            project = line.split("/")[-1]
            print(project, url + project)


if __name__ == "__main__":
    opts = parser.parse_args()
    for uri in opts.ssh:
        get_ssh_repos(uri)
    for user in opts.users:
        get_repos(login=user, type="user")
    for org in opts.orgs:
        get_repos(login=org, type="organization")
