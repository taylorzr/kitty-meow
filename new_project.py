import argparse
import json
import os
import re
import subprocess
from typing import List

from kitty.boss import Boss

parser = argparse.ArgumentParser(description="meow")

parser.add_argument(
    "--dir",
    dest="dirs",
    action="append",
    default=[],
    # required=True,
    help="directories to find projects",
)


# TODO: handle non github urls like
# https://src.fedoraproject.org/rpms/kitty.git
def main(args: List[str]) -> str:
    try:
        url = input("🐈 New project\nenter name or github url: ")
        return url
    except KeyboardInterrupt:
        return ""


def handle_result(
    args: List[str], answer: str, target_window_id: int, boss: Boss
) -> None:
    opts = parser.parse_args(args[1:])

    # This is the dir we clone repos into, for me it's not a big deal if they get cloned to the
    # first dir. But some people might want to pick which dir to clone to? How could that be
    # supported?
    projects_root = opts.dirs[0]

    if not answer:
        return
    elif "/" in answer:
        # Note: This is an attempt to see if the answer is a git url or not, e.g.
        #   - git@github.com:taylorzr/kitty-meow.git
        #   - https://github.com/taylorzr/kitty-meow.git
        github_url = answer
        dir = re.split("[/.]", github_url)[2]
        print(f"cloning into {dir}...")
        path = f"{projects_root}/{dir}"
        subprocess.run(["git", "clone", github_url, path])
    else:
        new_local = answer
        dir = new_local
        path = f"{projects_root}/{dir}"
        os.makedirs(path, exist_ok=True)

    # with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history", "a") as history:
    #     history.write(f"{dir} {datetime.now().isoformat()}\n")
    #     history.close()

    kitty_ls = json.loads(boss.call_remote_control(None, ("ls",)))
    for tab in kitty_ls[0]["tabs"]:
        # TODO: some sort of warning/message that the tab already existed
        if tab["title"] == dir:
            boss.call_remote_control(None, ("focus-tab", "--match", f"title:^{dir}$"))
            return

    # TODO: some sort of warning/message when the dir already exists
    window = boss.call_remote_control(
        None,
        (
            "launch",
            "--type",
            "tab",
            "--tab-title",
            dir,
            "--cwd",
            f"{path}",
            "zsh",
            "--interactive",
            "-c",
            "nvim",
        ),
    )

    boss.call_remote_control(
        window, ("launch", "--type", "window", "--dont-take-focus", "--cwd", "current")
    )
