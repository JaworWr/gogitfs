import pathlib
import tempfile
from dataclasses import dataclass

import git
import pytest
from _pytest.nodes import Item
from _pytest.reports import CollectReport
from _pytest.runner import CallInfo

from test.repo import Repo, load_repo_schema, build_repo

REPO_JSON = pathlib.Path(__file__).resolve().parent / "repo.json"


phase_report_key = pytest.StashKey[dict[str, pytest.CollectReport]]()


@pytest.hookimpl(wrapper=True, tryfirst=True)
def pytest_runtest_makereport(item: Item, call: CallInfo) -> CollectReport:
    # execute all other hooks to obtain the report object
    rep: CollectReport = yield

    # store test results for each phase of a call, which can
    # be "setup", "call", "teardown"
    item.stash.setdefault(phase_report_key, {})[rep.when] = rep

    return rep


@dataclass
class RepoInfo:
    path: pathlib.Path
    obj: git.Repo
    schema: Repo


@pytest.fixture(scope="session")
def repo():
    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = pathlib.Path(tmpdir)
        schema = load_repo_schema(REPO_JSON)
        repo = build_repo(schema, tmpdir)
        yield RepoInfo(path=tmpdir, obj=repo, schema=schema)


@pytest.fixture(scope="session")
def repo_path(repo):
    return repo.path


@pytest.fixture(scope="session")
def repo_obj(repo):
    return repo.obj


@pytest.fixture(scope="session")
def repo_schema(repo):
    return repo.schema


