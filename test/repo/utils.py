import os
from pathlib import Path

import git

from test.repo import schema, resolve


def load_repo(path: str | os.PathLike[str]) -> schema.Repo:
    with open(path) as f:
        repo_json = f.read()
    repo = schema.Repo.from_json(repo_json)
    return repo


def build_repo(repo_schema: schema.Repo, repo_path: str | os.PathLike[str]) -> git.Repo:
    repo_path = Path(repo_path)
    repo = git.Repo.init(repo_path, initial_branch=repo_schema.main_branch)

    repo_graph = make_graph_for_repo_schema(repo_schema)
    commit_order = resolve.resolve_graph(repo_graph)

    for commit_id in commit_order:
        branch = commit_id.split(":")[0]
        if branch != repo.active_branch.name:
            from_commit_id = repo_schema.branches[branch].from_commit
            if from_commit_id is not None:
                from_commit_hash = get_commit_hash(get_commit_by_id(repo_schema, from_commit_id))
            else:
                from_commit_hash = None
            checkout_branch(repo, branch, from_commit_hash)
        commit_schema = get_commit_by_id(repo_schema, commit_id)
        if isinstance(commit_schema, schema.Commit):
            make_commit(repo, repo_path, commit_schema)
        else:
            make_merge_commit(repo, repo_schema, commit_schema)
    checkout_branch(repo, repo_schema.active_branch)
    return repo


def make_graph_for_repo_schema(repo_schema: schema.Repo) -> resolve.Graph:
    graph = {}
    for branch_name, branch in repo_schema.branches.items():
        if branch_name != repo_schema.main_branch and branch.from_commit is None:
            raise RuntimeError(f"Only main branch ({repo_schema.main_branch}) can have from_commit == None")
        for idx, commit in enumerate(branch.commits):
            commit_id = f"{branch_name}:{idx}"
            parents = []
            if idx == 0:
                if branch.from_commit is not None:
                    parents.append(branch.from_commit)
            else:
                parents.append(f"{branch_name}:{idx-1}")
            if isinstance(commit, schema.MergeCommit):
                if not parents:
                    raise RuntimeError(f"Merge commit with only one parent: {commit.message}")
                parents.append(commit.other_commit)
            graph[commit_id] = parents
    return graph


def checkout_branch(repo: git.Repo, branch: str, branch_ref: str | None = None) -> git.Head:
    if branch not in repo.heads:
        if branch_ref is None:
            raise RuntimeError(f"Branch {branch} doesn't exist and hash wasn't provided")
        repo.create_head(branch, branch_ref)

    repo.heads[branch].checkout()
    return repo.heads[branch]


def make_commit_file(repo_path: Path, file_schema: schema.CommitFile) -> str:
    file_path = repo_path / file_schema.path
    file_path.parent.mkdir(exist_ok=True, parents=True)
    with open(file_path, "w") as f:
        f.write(file_schema.contents)
    return file_schema.path


def make_commit(repo: git.Repo, repo_path: Path, commit_schema: schema.Commit) -> git.Commit:
    for file_schema in commit_schema.files:
        file_path = make_commit_file(repo_path, file_schema)
        repo.index.add(file_path)
    commit = repo.index.commit(commit_schema.message, author_date=commit_schema.time)
    commit_schema.hash = commit.hexsha
    return commit


def make_merge_commit(
        repo: git.Repo,
        repo_schema: schema.Repo,
        commit_schema: schema.MergeCommit,
) -> git.Commit:
    other_commit_schema = get_commit_by_id(repo_schema, commit_schema.other_commit)
    other_hash = get_commit_hash(other_commit_schema)
    head = repo.head
    other_commit = repo.commit(other_hash)
    merge_base = repo.merge_base(head, other_commit)
    repo.index.merge_tree(other_commit, base=merge_base)
    merge_commit = repo.index.commit(
        message=commit_schema.message,
        author_date=commit_schema.time,
        parent_commits=[head.commit, other_commit],
    )
    commit_schema.hash = merge_commit.hexsha
    return merge_commit


def get_commit_hash(commit_schema: schema.Commit | schema.MergeCommit) -> str:
    h = commit_schema.hash
    if h is None:
        raise RuntimeError(f"Commit {commit_schema.message} not yet created")
    return h


def get_commit_by_id(repo_schema: schema.Repo, id_: str) -> schema.Commit | schema.MergeCommit:
    branch, idx = id_.split(":")
    idx = int(idx)
    return repo_schema.branches[branch].commits[idx]
