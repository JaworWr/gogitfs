from pathlib import Path

import git
import pytest

from test.repo import schema, utils

REPO_PATH = Path(__file__).resolve().parent / "small_repo.json"


@pytest.fixture
def small_repo_schema() -> schema.Repo:
    return utils.load_repo(REPO_PATH)


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


def test_get_commit_by_id(small_repo_schema: schema.Repo):
    commit_schema = utils.get_commit_by_id(small_repo_schema, "baz:1")
    assert commit_schema.message == "baz2"


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
    assert commit_schema.hash is not None
    assert commit.hexsha == commit_schema.hash
    assert commit.authored_date == commit_schema.time.timestamp()
    assert commit.message == commit_schema.message
    for f in commit_schema.files:
        assert (tmp_path / f.path).exists()


def test_checkout(tmp_path: Path):
    repo = git.Repo.init(tmp_path, initial_branch="main")
    repo.index.commit("foo1")
    assert sorted(h.name for h in repo.heads) == ["main"]
    utils.checkout_branch(repo, "branch", "HEAD")
    assert repo.active_branch.name == "branch"
    assert sorted(h.name for h in repo.heads) == ["branch", "main"]
    utils.checkout_branch(repo, "main")
    assert repo.active_branch.name == "main"
    assert sorted(h.name for h in repo.heads) == ["branch", "main"]
    utils.checkout_branch(repo, "branch")
    assert repo.active_branch.name == "branch"
    assert sorted(h.name for h in repo.heads) == ["branch", "main"]


def test_merge_commit(small_repo_schema: schema.Repo, tmp_path: Path):
    repo = git.Repo.init(tmp_path, initial_branch="main")
    main_commits = small_repo_schema.branches["main"].commits
    assert isinstance(main_commits[0], schema.Commit), "selected commit should not be a merge commit"
    utils.make_commit(repo, tmp_path, main_commits[0])
    assert isinstance(main_commits[1], schema.Commit), "selected commit should not be a merge commit"
    utils.make_commit(repo, tmp_path, main_commits[1])

    commit_schema = main_commits[2]
    assert isinstance(commit_schema, schema.MergeCommit), "selected commit should be a merge commit"
    branch = commit_schema.other_commit.split(":")[0]

    start_commit = utils.get_commit_by_id(small_repo_schema, small_repo_schema.branches[branch].from_commit)
    h = utils.get_commit_hash(start_commit)
    utils.checkout_branch(repo, branch, h)

    other_commit = small_repo_schema.branches[branch].commits[-1]
    utils.make_commit(repo, tmp_path, other_commit)
    utils.checkout_branch(repo, "main")
    commit = utils.make_merge_commit(repo, small_repo_schema, commit_schema)

    assert commit_schema.hash is not None
    assert commit.hexsha == commit_schema.hash
    assert commit.authored_date == commit_schema.time.timestamp()
    assert commit.message == commit_schema.message
    parent_hashes = [c.hexsha for c in commit.parents]
    expected_hahses = [main_commits[1].hash, other_commit.hash]
    assert sorted(parent_hashes) == sorted(expected_hahses)


def test_build_repo(small_repo_schema: schema.Repo, tmp_path: Path):
    repo = utils.build_repo(small_repo_schema, tmp_path)
    assert repo.active_branch.name == small_repo_schema.active_branch
    assert sorted(h.name for h in repo.heads) == sorted(small_repo_schema.branches)

    repo_commits = [c.hexsha for c in repo.iter_commits()]
    commit_schemas = [
        c.hash for c in small_repo_schema.branches["main"].commits + small_repo_schema.branches["bar"].commits
    ]
    assert sorted(repo_commits) == sorted(commit_schemas)

    for c in commit_schemas:
        if not isinstance(c, schema.Commit):
            continue
        for f in c.files:
            assert (tmp_path / f.path).exists(), f"{f.path} should exist"
            assert (tmp_path / f.path).read_text() == f.contents, f"{f.path} contents should match"

    repo.heads["baz"].checkout()
    repo_commits = [c.hexsha for c in repo.iter_commits()]
    commit_schemas = [
        c.hash for c in small_repo_schema.branches["main"].commits[:2] + small_repo_schema.branches["baz"].commits
    ]
    assert sorted(repo_commits) == sorted(commit_schemas)
