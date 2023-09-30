from pathlib import Path

import pytest

from test.repo import schema, utils

REPO_PATH = Path(__file__).resolve().parent / "small_repo.json"


@pytest.fixture
def small_repo_schema() -> schema.Repo:
    with open(REPO_PATH) as f:
        repo_json = f.read()
    repo_schema = schema.Repo.schema().loads(repo_json)
    return repo_schema


def test_repo_graph(small_repo_schema):
    graph = utils.make_graph_for_repo_schema(small_repo_schema)
    expected = {
        "main:0": [],
        "main:1": ["main:0", "bar:0"],
        "bar:0": ["main:0"],
        "baz:0": ["main:1"],
        "baz:1": ["baz:0"],
    }
    assert graph == expected
