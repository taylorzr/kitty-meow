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
    help="directories to find projects",
)

parser.add_argument(
    "--dirfile",
    dest="dirfile",
    help="file with directories to find projects",
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

def get_dirs_from_file(file):
    with open(file, "r") as f:
        return [line.strip() for line in f.readlines()]


def main(args: List[str]) -> tuple[str, bool]:
    opts = parser.parse_args(args[1:])

    # FIXME: How to call boss in the main function?
    # data = boss.call_remote_control(None, ("ls",))
    kitty_ls = json.loads(
        subprocess.run(
            ["kitty", "@", "ls"], capture_output=True, text=True
        ).stdout.strip("\n")
    )

    tabs = [tab["title"] for tab in kitty_ls[0]["tabs"]]
    tabs_and_projects = [tab["title"] for tab in kitty_ls[0]["tabs"]]
    projects = []

    if opts.dirfile:
        opts.dirs += [line for line in get_dirs_from_file(opts.dirfile) if line not in opts.dirs]

    for dir in opts.dirs:
        dir = os.path.expanduser(dir)
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

    default_prompt = "🐈project"
    flags = []
    for org in opts.orgs:
        flags.append(f"--org {org}")
    for user in opts.users:
        flags.append(f"--user {user}")
    # NOTE: Can't use ' char within any of the binds
    binds = [
        f"ctrl-r:change-prompt(🐈remote> )+reload({bin_path}python3 ~/.config/kitty/meow/get_all_repos.py {' '.join(flags)})",
        'ctrl-t:change-prompt(🐈tabs> )+reload(printf "{0}")'.format("\n".join(tabs)),
        'alt-p:change-prompt(🐈projects> )+reload(printf "{0}")'.format(
            "\n".join(projects)
        ),
        'alt-l:change-prompt({0}> )+reload(printf "{1}")'.format(
            default_prompt, "\n".join(tabs_and_projects)
        ),
    ]
    args = [
        f"{bin_path}fzf",
        f"--header=ctrl-r: remote | alt-p: project | ctrl-t: tabs | alt-l: tabs&projects",
        f"--prompt={default_prompt}> ",
        f"--bind={','.join(binds)}",
    ]
    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(tabs_and_projects).encode())[0]
    selection = out.decode().strip()

    # from kittens.tui.loop import debug
    # debug(selection)

    as_project = False

    if len(selection) and selection not in tabs:
        as_project = input(f'Start {selection} in project mode? [y,N] >')
        if as_project.upper() == 'Y':
            as_project = True
        else:
            as_project = False

    return (selection, as_project)


def handle_result(
    args: List[str], answer: tuple[str, bool], target_window_id: int, boss: Boss
) -> None:
    opts = parser.parse_args(args[1:])

    # This is the dir we clone repos into, for me it's not a big deal if they get cloned to the
    # first dir. But some people might want to pick which dir to clone to? How could that be
    # supported?

    # TODO: make it possilbe to mnto fail if only --dirfile is used
    if opts.dirs:
        projects_root = opts.dirs[0]
    else:
        projects_root = get_dirs_from_file(opts.dirfile)[0]

    (path, as_project) = answer
    path, *rest = path.split()
    dir = os.path.basename(path)

    if len(rest) == 1:
        ssh_url = rest[0]
        print(f"cloning into {dir}...")
        path = f"{projects_root}/{dir}"
        subprocess.run(["git", "clone", ssh_url, path])
        # TODO: handle error, like unset sso on ssh key and try this
    elif len(rest) != 0:
        print("something bad happenend :(")

    with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history", "a") as history:
        history.write(f"{dir} {datetime.now().isoformat()}\n")
        history.close()

    kitty_ls = json.loads(boss.call_remote_control(None, ("ls",)))
    for tab in kitty_ls[0]["tabs"]:
        if tab["title"] == dir:
            boss.call_remote_control(None, ("focus-tab", "--match", f"title:^{dir}$"))
            return

    window_id = boss.call_remote_control(
        None,
        (
            "launch",
            "--type",
            "tab",
            "--tab-title",
            dir,
            "--cwd",
            path,
        ),
    )

    parent_window = boss.window_id_map.get(window_id)

    # start editor and another window if as_project is True
    if as_project:
        boss.call_remote_control(parent_window, ("send-text", "${EDITOR:-vim}\n"))
        boss.call_remote_control(
            parent_window,
            ("launch", "--type", "window", "--dont-take-focus", "--cwd", path),
        )
