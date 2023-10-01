import pathlib
import tempfile
from dataclasses import dataclass

from test.repo import Repo, load_repo_schema, build_repo

import pytest

REPO_JSON = pathlib.Path(__file__).resolve().parent / "repo.json"


@dataclass
class RepoInfo:
    path: pathlib.Path
    schema: Repo


@pytest.fixture(scope="session")
def repo():
    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = pathlib.Path(tmpdir)
        schema = load_repo_schema(REPO_JSON)
        build_repo(schema, tmpdir)
        yield RepoInfo(path=tmpdir, schema=schema)


@pytest.fixture(scope="session")
def repo_path(repo):
    return repo.path


@pytest.fixture(scope="session")
def repo_schema(repo):
    return repo.schema


