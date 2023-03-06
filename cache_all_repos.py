import argparse
import os
import sys
from typing import List

sys.path.append("/opt/homebrew/lib/python3.10/site-packages")
sys.path.append("~/.config/kitty/meow")

import github

parser = argparse.ArgumentParser(description="meow")

parser.add_argument(
    "--org",
    dest="orgs",
    action="append",
    default=[],
    help="look for repos in these github orgs",
)


def main(args: List[str]) -> str:
    opts = parser.parse_args(args[1:])

    # TODO: Gotta think through caching with multi-org
    for org in opts.orgs:
        cache = f"{os.path.expanduser('~')}/.config/kitty/meow/cache_{org}"
        repos = github.get_all_repos(org)
        with open(cache, "w") as file:
            file.write("\n".join(repos))
