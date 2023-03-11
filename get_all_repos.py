import argparse
import os
import sys
from sys import stdout

import requests

parser = argparse.ArgumentParser(description="meow")

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

# TODO: Also get repos for current user and merge them together with any org
token = os.getenv("GITHUB_TOKEN")

org_query = """
query($login: String!, $cursor: String) {
    organization(login: $login) {
        repositories(first: 100, after: $cursor) {
            nodes {
                name
                sshUrl
                }
            pageInfo {
                endCursor
                startCursor
                hasNextPage
            }
        }
    }
}
"""

user_query = """
query($login: String!, $cursor: String) {
    user(login: $login) {
        repositories(first: 100, after: $cursor) {
            nodes {
               name
                sshUrl
                }
            pageInfo {
                endCursor
                startCursor
                hasNextPage
            }
        }
    }
}
"""


def run_query(
    query, login, cursor=None
):
    request = requests.post(
        "https://api.github.com/graphql",
        json={"query": query, "variables": {"login": login, "cursor": cursor}},
        headers={"Authorization": f"Bearer {token}"},
    )
    if request.status_code == 200:
        return request.json()
    else:
        raise Exception(
            "Query failed to run by returning code of {}. {}".format(
                request.status_code, query
            )
        )

def get_repos(login, type):
    if type == "organization":
        query = org_query
    elif type == "user":
        query = user_query
    else:
        error

    cursor = None
    repos = []

    while True:
        result = run_query(query, login, cursor)

        for repo in result["data"][type]["repositories"]["nodes"]:
            print(repo["name"], repo["sshUrl"])
        stdout.flush()  # so that results show in fzf sooner

        pageInfo = result["data"][type]["repositories"]["pageInfo"]
        if not pageInfo["hasNextPage"]:
            break
        cursor = pageInfo["endCursor"]

    return repos

if __name__ == "__main__":
    opts = parser.parse_args()
    for user in opts.users:
        get_repos(login=user, type="user")
    for org in opts.orgs:
        get_repos(login=org, type="organization")
