import os
import json
from sys import stdout
from urllib.request import Request, urlopen

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


def run_query(query, login, cursor=None):
    data = {"query": query, "variables": {"login": login, "cursor": cursor}}
    headers = {"Authorization": f"Bearer {token}"}
    request = Request(
        "https://api.github.com/graphql", json.dumps(data).encode("utf-8"), headers
    )
    with urlopen(request) as response:
        code = response.code
        body = response.read()

    if code == 200:
        data = json.loads(body)
        if "errors" in data:
            raise Exception(f'Query returned errors:\n{data["errors"]}')
        return data
    else:
        raise Exception(
            "Query failed to run by returning code of {}. {}".format(code, query)
        )


def get_repos(login, type):
    if type == "organization":
        query = org_query
    elif type == "user":
        query = user_query
    else:
        raise RuntimeError(
            f"expected type to be `user` or `organization`, but got `{type}`"
        )

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
