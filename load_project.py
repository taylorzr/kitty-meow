import argparse
import json
import os
import subprocess
import sys
from typing import List

from kitty.boss import Boss

parser = argparse.ArgumentParser(description="meow")

parser.add_argument(
    "--dir",
    dest="dirs",
    action="append",
    default=[],
    # required=True,
    help="direcories to find projects",
)

parser.add_argument(
    "--org",
    dest="orgs",
    action="append",
    default=[],
    help="look for repos in these github orgs",
)


def main(args: List[str]) -> str:
    opts = parser.parse_args(args[1:])
    org = opts.orgs[0]

    projects = []
    for dir in opts.dirs:
        if dir.endswith("/"):
            for f in os.scandir(dir):
                if f.is_dir():
                    projects.append(f.path)
        else:
            projects.append(dir)

    # from kittens.tui.loop import debug
    # debug(projects)

    # FIXME: How to call boss in the main function?
    # data = boss.call_remote_control(None, ("ls",))
    stuff = subprocess.run(
        ["kitty", "@", "ls"], capture_output=True, text=True
    ).stdout.strip("\n")
    data = json.loads(stuff)

    bin_path = os.getenv("BIN_PATH", "")

    # TODO: Cache github results, can i refresh async somehow?
    # Or can I have a binding that forces refresh?
    # Or just bg spawn a function to get and write all repos to cache every time run

    # NOTE: Can't use ' char within any of the binds
    # TODO: bind ctrl-x to kill a tab using fzf execute
    # ^ did this, but need to fix reload tabs afterwards
    # or maybe make a separate command for that?
    binds = [
        f'ctrl-r:change-prompt(local> )+reload(ls -d1 {" ".join(projects)} | sed "s|{os.path.expanduser("~")}|~|")',
        f"ctrl-g:change-prompt(github> )+reload({bin_path}python3 ~/.config/kitty/meow/get_all_repos.py {org})",
        'ctrl-x:execute(kitty @ close-tab --match=title:{})+reload(kitty @ ls | jq -r ".[0].tabs | map(.title) | .[]")',
    ]

    args = [f"{bin_path}fzf", "--prompt=tabs> ", f"--bind={','.join(binds)}"]

    tabs = [tab["title"] for tab in data[0]["tabs"]]

    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(tabs).encode())[0]
    selection = out.decode().strip().encode()

    if len(selection) > 0:
        return selection[0]


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

    path, *rest = answer.split()
    dir = os.path.basename(path)

    if len(rest) == 1:
        ssh_url = rest[0]
        print(f"cloning into {dir}...")
        path = f"{projects_root}/{dir}"
        subprocess.run(["git", "clone", ssh_url, path])
    elif len(rest) != 0:
        print("something bad happenend :(")

    stuff = boss.call_remote_control(None, ("ls",))
    data = json.loads(stuff)

    for tab in data[0]["tabs"]:
        if tab["title"] == dir:
            boss.call_remote_control(None, ("focus-tab", "--match", f"title:{dir}"))
            return

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
            "-lic",
            "nvim",
        ),
    )
    boss.call_remote_control(
        window, ("launch", "--type", "window", "--dont-take-focus", "--cwd", "current")
    )
