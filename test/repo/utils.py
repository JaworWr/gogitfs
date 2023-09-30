import os
from pathlib import Path

import git

from test.repo import schema


def load_repo(path: str | os.PathLike[str]) -> schema.Repo:
    with open(path) as f:
        repo_json = f.read()
    repo = schema.Repo.schema().loads(repo_json)
    return repo


def build_repo(repo_schema: schema.Repo, repo_path: str | os.PathLike[str]) -> git.Repo:
    repo_path = Path(repo_path)
    repo = git.Repo.init(repo_path)
    return repo


def checkout_branch(repo: git.Repo, branch: str, branch_hash: str | None = None) -> git.Head:
    if branch not in repo.heads:
        if branch_hash is None:
            raise RuntimeError(f"Branch {branch} doesn't exist and hash wasn't provided")
        repo.create_head(branch, branch_hash)

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
        parent_commits=[head.commit, other_commit],
    )
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
