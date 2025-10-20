import argparse
import os
import subprocess
from enum import Enum
from urllib.parse import urlparse

import github

parser = argparse.ArgumentParser(description="meow")


class Modes(Enum):
    git = "git"
    ssh = "ssh"


parser.add_argument("mode", type=Modes, choices=list(Modes))

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


def print_repos(login, type):
    cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{login}"
    try:
        with open(cache, "r") as file:
            print(file.read())
    except FileNotFoundError:
        github.print_repos(login, type)


def print_ssh_repos(url):
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

    # TODO: error if we don't have the parts we need, at least a dest & path

    if not uri.path.endswith("/"):
        project = url.split("/")[-1]
        print(project, url)
        return

    result = subprocess.run(
        f"ssh -o ConnectTimeout=5 {dest} 'ls -d {uri.path.lstrip("/")}*'",
        shell=True,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print("Error:", result.stderr.rstrip())
    else:
        for line in result.stdout.splitlines():
            if line:
                project = line.split("/")[-1]
                print(project, url + project)


if __name__ == "__main__":
    opts = parser.parse_args()

    if opts.mode == Modes.ssh:
        for uri in opts.ssh:
            print_ssh_repos(uri)

    if opts.mode == Modes.git:
        for user in opts.users:
            print_repos(login=user, type="user")
        for org in opts.orgs:
            print_repos(login=org, type="organization")
