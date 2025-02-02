import json
import os
import subprocess
from datetime import datetime, timedelta
from typing import List

from kitty.boss import Boss

import meow


def main(args: List[str]) -> str:
    last_views = {}

    with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history") as history:
        for line in history:
            parts = line.strip().split(" ")
            if len(parts) != 2:
                print(
                    f"expected history to have 2 parts (project last-time), but found {len(parts)}"
                )
                continue
            last_views[parts[0]] = parts[1]

    # TODO: make time configurable
    cutoff = datetime.now() - timedelta(days=3)

    stuff = subprocess.run(
        ["kitty", "@", "ls"], capture_output=True, text=True
    ).stdout.strip("\n")
    data = json.loads(stuff)

    all_tabs, old_tabs = [], []

    for tab in data[0]["tabs"]:
        title = tab["title"]
        all_tabs.append(title)

        last_view = last_views.get(title, None)
        if not last_view or datetime.fromisoformat(last_view) < cutoff:
            old_tabs.append(title)

    bin_path = os.getenv("BIN_PATH", "")

    binds, header = meow.binds_and_header(
        {
            "ctrl-o": (
                "old",
                'printf "{0}"'.format("\n".join(old_tabs)),
            ),
            "ctrl-a": (
                "any",
                'printf "{0}"'.format("\n".join(all_tabs)),
            ),
        }, emoji="🐈💀"
    )
    args = [
        f"{bin_path}fzf",
        "--multi",
        "--reverse",
        f"--header={header}",
        f"--bind={binds}",
        f"--prompt=🐈💀 kill > ",
    ]
    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(all_tabs).encode())[0]
    selections = out.decode().strip()

    if selections == "":
        return

    for tab in selections.split("\n"):
        subprocess.run(["kitty", "@", "close-tab", "--match", f"title:{tab}"])


def handle_result(
    args: List[str], answer: str, target_window_id: int, boss: Boss
) -> None:
    pass
