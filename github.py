import os
from sys import stdout

import requests

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
        raise RuntimeError(f"expected type to be `user` or `organization`, but got `{type}`")

    cursor = None
    repos = []

    while True:
        result = run_query(query, login, cursor)

        for repo in result["data"][type]["repositories"]["nodes"]:
            output = f'{repo["name"]} {repo["sshUrl"]}'
            repos.append(output)
            print(output)
        stdout.flush()  # so that results show in fzf sooner

        pageInfo = result["data"][type]["repositories"]["pageInfo"]
        if not pageInfo["hasNextPage"]:
            break
        cursor = pageInfo["endCursor"]

    return repos
