import os
from sys import stdout

import requests

# TODO: Also get repos for current user and merge them together with any org
token = os.getenv("GITHUB_TOKEN")

query = """
query($org: String!, $cursor: String) {
    organization(login: $org) {
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


def run_query(query, org, cursor=None):
    request = requests.post(
        "https://api.github.com/graphql",
        json={"query": query, "variables": {"org": org, "cursor": cursor}},
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


def get_all_repos(org):
    cursor = None

    repos = []
    while True:
        result = run_query(query, org=org, cursor=cursor)

        for repo in result["data"]["organization"]["repositories"]["nodes"]:
            output = f'{repo["name"]} {repo["sshUrl"]}'
            repos.append(output)
            print(output)
        stdout.flush()  # so that results show in fzf sooner

        pageInfo = result["data"]["organization"]["repositories"]["pageInfo"]
        if not pageInfo["hasNextPage"]:
            break
        cursor = pageInfo["endCursor"]

    return repos
