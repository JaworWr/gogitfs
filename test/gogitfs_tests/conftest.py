import pathlib
import tempfile
from dataclasses import dataclass

import git

from test.repo import Repo, load_repo_schema, build_repo

import pytest

REPO_JSON = pathlib.Path(__file__).resolve().parent / "repo.json"


@dataclass
class RepoInfo:
    path: pathlib.Path
    repo_object: git.Repo
    schema: Repo


@pytest.fixture(scope="session")
def repo():
    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = pathlib.Path(tmpdir)
        schema = load_repo_schema(REPO_JSON)
        repo = build_repo(schema, tmpdir)
        yield RepoInfo(path=tmpdir, repo_object=repo, schema=schema)


@pytest.fixture(scope="session")
def repo_path(repo):
    return repo.path


@pytest.fixture(scope="session")
def repo_object(repo):
    return repo.repo_object


@pytest.fixture(scope="session")
def repo_schema(repo):
    return repo.schema


