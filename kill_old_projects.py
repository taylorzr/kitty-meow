import json
import os
import subprocess
from datetime import datetime, timedelta
from typing import List

from kitty.boss import Boss


def main(args: List[str]) -> str:
    last_views = {}

    with open(f"{os.path.expanduser('~')}/.config/kitty/meow/history") as history:
        for line in history:
            parts = line.strip().split(" ")
            if len(parts) != 2:
                import code

                code.interact(local=dict(globals(), **locals()))
            last_views[parts[0]] = parts[1]

    # TODO: make time configurable
    cutoff = datetime.now() - timedelta(days=3)

    stuff = subprocess.run(
        ["kitty", "@", "ls"], capture_output=True, text=True
    ).stdout.strip("\n")
    data = json.loads(stuff)

    old_tabs = []

    for tab in data[0]["tabs"]:
        title = tab["title"]
        last_view = last_views.get(title, None)

        if not last_view or datetime.fromisoformat(last_view) < cutoff:
            # print(f"{title} {last_view or '-'}")
            old_tabs.append(title)

    bin_path = os.getenv("BIN_PATH", "")

    args = [f"{bin_path}fzf", "--prompt=kill> ", "--multi"]
    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    out = p.communicate(input="\n".join(old_tabs).encode())[0]
    selections = out.decode().strip()

    if selections == "":
        return

    for tab in selections.split("\n"):
        subprocess.run(["kitty", "@", "close-tab", "--match", f"title:{tab}"])


def handle_result(
    args: List[str], answer: str, target_window_id: int, boss: Boss
) -> None:
    pass
