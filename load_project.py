import argparse
import json
import os
import subprocess
from datetime import datetime
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

parser.add_argument(
    "--user",
    dest="users",
    action="append",
    default=[],
    help="look for repos for these github users",
)


def main(args: List[str]) -> str:
    opts = parser.parse_args(args[1:])

    # FIXME: How to call boss in the main function?
    # data = boss.call_remote_control(None, ("ls",))
    stuff = subprocess.run(
        ["kitty", "@", "ls"], capture_output=True, text=True
    ).stdout.strip("\n")
    data = json.loads(stuff)

    tabs = [tab["title"] for tab in data[0]["tabs"]]
    tabs_and_projects = [tab["title"] for tab in data[0]["tabs"]]
    projects = []

    for dir in opts.dirs:
        if dir.endswith("/"):
            for f in os.scandir(dir):
                if f.is_dir():
                    name = os.path.basename(f.path)
                    pretty_path = f.path.replace(os.path.expanduser("~"), "~", 1)
                    projects.append(pretty_path)
                    if name not in tabs_and_projects:
                        tabs_and_projects.append(pretty_path)
        else:
            name = os.path.basename(dir)
            projects.append(dir)
            if name not in tabs_and_projects:
                tabs_and_projects.append(dir)

    bin_path = os.getenv("BIN_PATH", "")

    # TODO: Cache github results, can i refresh async somehow?
    # Or can I have a binding that forces refresh?
    # Or just bg spawn a function to get and write all repos to cache every time run

    default_prompt = "tabs&projects"
    flags = []
    for org in opts.orgs:
        flags.append(f"--org {org}")
    for user in opts.users:
        flags.append(f"--user {user}")
    # NOTE: Can't use ' char within any of the binds
    binds = [
        'ctrl-r:change-prompt({0}> )+reload(print "{1}")'.format(default_prompt, "\n".join(tabs_and_projects)),
        'ctrl-t:change-prompt(tabs> )+reload(print "{0}")'.format("\n".join(tabs)),
        'ctrl-p:change-prompt(projects> )+reload(print "{0}")'.format("\n".join(projects)),
        f"ctrl-g:change-prompt(github> )+reload({bin_path}python3 ~/.config/kitty/meow/get_all_repos.py {' '.join(flags)})",
    ]
    args = [f"{bin_path}fzf", f"--prompt={default_prompt}> ", f"--bind={','.join(binds)}"]
    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(tabs_and_projects).encode())[0]
    selection = out.decode().strip()

    # from kittens.tui.loop import debug
    # debug(selection)

    return selection


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

    with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history", "a") as history:
        history.write(f"{dir} {datetime.now().isoformat()}\n")
        history.close()

    for tab in data[0]["tabs"]:
        if tab["title"] == dir:
            boss.call_remote_control(None, ("focus-tab", "--match", f"title:^{dir}$"))
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
            "--interactive",
            "-c",
            "nvim",
        ),
    )
    boss.call_remote_control(
        window, ("launch", "--type", "window", "--dont-take-focus", "--cwd", "current")
    )
