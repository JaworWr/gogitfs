import os
from pathlib import Path

import git

from test.repo import schema


def load_repo(path: str | os.PathLike[str]) -> schema.Repo:
    with open(path) as f:
        repo_json = f.read()
    repo = schema.Repo.schema().loads(repo_json)
    return repo


def build_repo(repo_path: str | os.PathLike[str], repo_schema: schema.Repo) -> git.Repo:
    repo_path = Path(repo_path)
    repo = git.Repo.init(repo_path)
    return repo


def make_commit_file(repo_path: Path, file_schema: schema.CommitFile) -> str:
    file_path = repo_path / file_schema.path
    file_path.parent.mkdir(exist_ok=True, parents=True)
    with open(file_path, "w") as f:
        f.write(file_schema.contents)
    return file_schema.path


def make_commit(repo_path: Path, repo: git.Repo, commit_schema: schema.Commit) -> git.Commit:
    for file_schema in commit_schema.files:
        file_path = make_commit_file(repo_path, file_schema)
        repo.index.add(file_path)
    commit = repo.index.commit(commit_schema.message, author_date=commit_schema.time)
    commit_schema.hash = commit.hexsha
    return commit
