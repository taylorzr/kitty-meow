import json
import os
import subprocess
import sys
from typing import List

from kitty.boss import Boss

# fmt: off
sys.path.insert(0, "/opt/homebrew/lib/python3.10/site-packages")
from pyfzf.pyfzf import FzfPrompt

# fmt:on


def main(args: List[str]) -> str:
    org = args[1]
    projects_root = args[2].rstrip("/")
    dirs = [f.name for f in os.scandir(projects_root) if f.is_dir()]

    # FIXME: How to call boss in the main function?
    # data = boss.call_remote_control(None, ("ls",))
    stuff = subprocess.run(
        ["kitty", "@", "ls"], capture_output=True, text=True
    ).stdout.strip("\n")
    data = json.loads(stuff)

    # first os window, how to handle multiple os windows?
    tabs = [tab.title for tab in data[0]]

    non_open_projects = list(set(dirs) - set(tabs))

    # TODO: Cache github results, can i refresh async somehow?
    # Or can I have a binding that forces refresn?
    # Or just bg spawn a function to get and write all repos to cache every time run
    fzf = FzfPrompt("/opt/homebrew/bin/fzf")

    selection = fzf.prompt(
        non_open_projects,
        f"--bind 'ctrl-g:reload(/opt/homebrew/opt/python@3.10/libexec/bin/python ~/.config/kitty/meow/get_all_repos.py {org}),ctrl-r:reload(eval ls -1 ~/code)'",
    )

    if len(selection) > 0:
        return selection[0]


def handle_result(
    args: List[str], answer: str, target_window_id: int, boss: Boss
) -> None:
    projects_root = args[2].rstrip("/")

    if not answer:
        return

    dir, *rest = answer.split()

    if len(rest) == 1:
        ssh_url = rest[0]
        print(f"cloning into {dir}...")
        subprocess.run(["git", "clone", ssh_url, f"{projects_root}/{dir}"])
    elif len(rest) != 0:
        print("something bad happenend :(")

    window = boss.call_remote_control(
        None,
        (
            "launch",
            "--type",
            "tab",
            "--tab-title",
            dir,
            "--cwd",
            f"{projects_root}/{dir}",
            "zsh",
            "-lic",
            "nvim",
        ),
    )
    boss.call_remote_control(
        window, ("launch", "--type", "window", "--dont-take-focus", "--cwd", "current")
    )
