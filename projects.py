from kittens.tui.handler import kitten_ui
from urllib.parse import urlparse
import argparse
import json
import os
import re
import subprocess
from datetime import datetime
from typing import List

from kitty.boss import Boss

import meow


@kitten_ui(allow_remote_control=True)
def main(args: List[str]) -> str:
    opts = parser.parse_args(args[1:])

    if opts.command == "load":
        return main_load(args, opts)
    elif opts.command == "new":
        return main_new(args, opts)
    else:
        raise ValueError(f"i don't know what do to with command {opts.command}")


def handle_result(
    args: List[str], answer: str, target_window_id: int, boss: Boss
) -> None:
    opts = parser.parse_args(args[1:])

    if opts.command == "load":
        return handle_load(args, answer, target_window_id, boss)
    elif opts.command == "new":
        return handle_new(args, answer, target_window_id, boss)


def main_load(args, opts):
    cp = main.remote_control(["ls"], capture_output=True)
    kitty_tabs = json.loads(cp.stdout.decode("utf-8").strip("\n"))[0]["tabs"]

    tabs = [tab["title"] for tab in kitty_tabs]
    tabs_and_projects = [tab["title"] for tab in kitty_tabs]

    projects = []
    for dir in opts.dirs:
        if dir.endswith("/"):
            for f in os.scandir(dir):
                if f.is_dir():
                    name = os.path.basename(f.path)
                    pretty_path = f.path.replace(os.path.expanduser("~"), "~", 1)
                    # TODO: can we use term escape codes to make the 2nd part lighter?
                    line = name + " " + pretty_path
                    projects.append(line)
                    if name not in tabs_and_projects:
                        tabs_and_projects.append(line)
        else:
            name = os.path.basename(dir)
            projects.append(dir)
            if name not in tabs_and_projects:
                tabs_and_projects.append(dir)

    bin_path = os.getenv("BIN_PATH", "")

    flags = []
    for org in opts.orgs:
        flags.append(f"--org {org}")
    for user in opts.users:
        flags.append(f"--user {user}")
    for host in opts.ssh:
        flags.append(f"--ssh {host}")

    # NOTE: don't use
    # - ctrl-p -> fzf previous item in list
    # - ctrl-n -> fzf next item in list
    binds, header = meow.binds_and_header(
        {
            "ctrl-t": (
                "tabs",
                'printf "{0}"'.format("\n".join(tabs)),
            ),
            "ctrl-o": (
                "local",
                'printf "{0}"'.format("\n".join(projects)),
            ),
            "ctrl-r": (
                "remote",
                f"{bin_path}python3 ~/.config/kitty/meow/get_all_repos.py {' '.join(flags)}",
            ),
            "ctrl-i": (
                "history",
                # TODO: make a version that shows uniqueness?
                f"{bin_path}tac ~/.config/kitty/meow/history",
            ),
            "ctrl-a": (
                "tabs&projects",
                'printf "{0}"'.format("\n".join(tabs_and_projects)),
            ),
        }
    )

    command = [
        f"{bin_path}fzf",
        "--multi",
        "--reverse",
        f"--header={header}",
        f"--bind={binds}",
        "--prompt=🐈 meow > ",
    ]
    p = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(tabs_and_projects).encode())[0]
    output = out.decode().strip()

    # from kittens.tui.loop import debug
    # debug(output)

    if output == "":
        return []

    selections = output.split("\n")

    stuff = []

    for selection in selections:
        # NOTE: selection can be a variety of patterns
        # - meow
        # - ~/.config/kitty/meow/
        # - kitty-meow git@github.com:taylorzr/kitty-meow.git
        # - meow 2023-03-23T20:52:21.841221
        # - homelab ssh://127.0.0.1/code/homelab
        name, *rest = selection.split()
        if len(rest) > 2:
            raise ValueError(f"expected <= 2 parts, a project name and optionally a uri, but got {len(rest)} parts")
        url = rest[0] if len(rest) > 0 else ""
        project = [name, url]

        uri = urlparse(url)
        # TODO: handle non-github urls
        if url and "git@github.com:" in uri.path and not uri.scheme == "ssh":
            opts = parser.parse_args(args[1:])
            projects_root = opts.dirs[0]
            # NOTE: We clone into the first --dir flag called. Does this need to be configurable,
            # how could we do that?
            clone_path = f"{projects_root}/{name}"
            cp = subprocess.run(["git", "clone", url, clone_path])
            # TODO: handle any cp errors
            project[1] = clone_path
            # TODO: handle error, like unset sso on ssh key and try this

        stuff.append(project)

    return stuff


def main_new(args, opts):
    try:
        url = input("🐈 new (name or github url): ")
        return url
    except KeyboardInterrupt:
        return ""


def handle_load(args: List[str], stuff, target_window_id: int, boss: Boss):
    if not stuff:
        return

    for (name, url) in stuff:
        load_project(boss, name, url)


def handle_new(args: List[str], answer: str, target_window_id: int, boss: Boss):
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
        name = re.split("[/.]", github_url)[2]
        print(f"cloning into {name}...")
        path = f"{projects_root}/{name}"
        subprocess.run(["git", "clone", github_url, path])
    else:
        new_local = answer
        name = new_local
        path = f"{projects_root}/{name}"
        os.makedirs(path, exist_ok=True)

    load_project(boss, name, path)


def load_project(boss, name, path_or_url):
    uri = urlparse(path_or_url)

    tab_title = name
    if uri.scheme == "ssh":
        tab_title = f"{name}-ssh"

    with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history", "a") as history:
        history.write(f"{tab_title} {datetime.now().isoformat()}\n")
        history.close()

    # NOTE: if project already open, just switch to it
    kitty_ls = json.loads(boss.call_remote_control(None, ("ls",)))
    for tab in kitty_ls[0]["tabs"]:
        if tab["title"] == tab_title:
            boss.call_remote_control(None, ("focus-tab", "--match", f"title:^{tab_title}$"))
            return

    # start editor and another window
    if uri.scheme == "ssh":
        path = uri.path.lstrip("/")
        window_id = boss.call_remote_control(None, (
            "launch", "--type", "tab", "--tab-title", tab_title,
            # FIX: for some reason, passing a command breaks starting new windows in ssh
            # FIX: start editor not vim
            "kitty", "+kitten", "ssh", "-t", uri.netloc, f"cd {path}; vim"
            # "kitty", "+kitten", "ssh", "192.168.1.10"
        ))
        parent_window = boss.window_id_map.get(int(window_id))
        # command.append("kitty", "+kitten", "ssh", "-t", "192.168.1.10", "cd code/homelab; vim")
        # FIX: for some reason, passing a command breaks starting new windows in ssh
        # FIX: start user shell not bash
        boss.call_remote_control(parent_window, ("launch", "kitten", "ssh", "-t", uri.netloc, f"cd {path}; bash"))
    else:
        window_id = boss.call_remote_control(None, (
            "launch",
            "--type",
            "tab",
            "--tab-title",
            name,
            "--title",
            name,
            "--cwd",
            path_or_url,
        ))
        parent_window = boss.window_id_map.get(int(window_id))
        boss.call_remote_control(parent_window, ("send-text", "${EDITOR:-vim}\n"))
        boss.call_remote_control(
            parent_window,
            ("launch", "--type", "window", "--title", "current", "--dont-take-focus", "--cwd", path_or_url),
            # FIX: title works for the first window, but children don't inherit
            # probably need to edit the splitter binding
        )


parser = argparse.ArgumentParser(description="meow")
parser.add_argument("command", nargs="?", default="load")
parser.add_argument(
    "--dir",
    dest="dirs",
    action="append",
    default=[],
    help="directories to find projects",
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
parser.add_argument(
    "--ssh",
    dest="ssh",
    action="append",
    default=[],
    help="look for projects in these ssh paths",
)
