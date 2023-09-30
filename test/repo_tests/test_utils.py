from pathlib import Path

import git
import pytest

from test.repo import schema, utils

REPO_PATH = Path(__file__).resolve().parent / "small_repo.json"


@pytest.fixture
def small_repo_schema() -> schema.Repo:
    with open(REPO_PATH) as f:
        repo_json = f.read()
    repo_schema = schema.Repo.schema().loads(repo_json)
    return repo_schema


def test_repo_graph(small_repo_schema: schema.Repo):
    graph = utils.make_graph_for_repo_schema(small_repo_schema)
    expected = {
        "main:0": [],
        "main:1": ["main:0"],
        "main:2": ["main:1", "bar:0"],
        "bar:0": ["main:0"],
        "baz:0": ["main:1"],
        "baz:1": ["baz:0"],
    }
    assert graph == expected


def test_commit_files(small_repo_schema: schema.Repo, tmp_path: Path):
    files = small_repo_schema.branches["main"].commits[0].files
    for f in files:
        utils.make_commit_file(tmp_path, f)
        assert (tmp_path / f.path).exists(), f"{f.path} should exist"
        assert (tmp_path / f.path).read_text() == f.contents, f"{f.path} contents should match"


def test_commit(small_repo_schema: schema.Repo, tmp_path: Path):
    commit_schema = small_repo_schema.branches["main"].commits[0]
    assert isinstance(commit_schema, schema.Commit), "selected commit should not be a merge commit"
    repo = git.Repo.init(tmp_path)
    commit = utils.make_commit(repo, tmp_path, commit_schema)
    assert commit.hexsha == commit_schema.hash
    assert commit.authored_date == commit_schema.time.timestamp()
    assert commit.message == commit_schema.message
    for f in commit_schema.files:
        assert (tmp_path / f.path).exists()
