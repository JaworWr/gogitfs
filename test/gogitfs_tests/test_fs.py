import os
import pathlib
import subprocess

import pytest
import sh

from test.gogitfs_tests.common import GOGITFS_BINARY
from test.repo import schema


@pytest.fixture
def mount(repo_path: pathlib.Path, tmp_path: pathlib.Path):
    args = [GOGITFS_BINARY, str(repo_path), str(tmp_path)]
    p = subprocess.Popen(args)
    p.wait()
    assert p.returncode == 0
    yield tmp_path
    sh.umount(tmp_path)


def test_mount_subdirs(mount: pathlib.Path):
    assert (mount / "commits").is_dir()
    assert (mount / "branches").is_dir()


def check_commit_base(commit_dir: pathlib.Path, commit: schema.Commit | schema.MergeCommit):
    assert (commit_dir / "hash").read_text() == commit.hash
    assert (commit_dir / "message").read_text() == commit.message
    for f in commit_dir.glob("**"):
        stat = os.stat(f)
        assert stat.st_mtime == stat.st_atime == stat.st_ctime == commit.time.timestamp()


def check_commit(repo_schema: schema.Repo, commit_dir: pathlib.Path, name: str, commit: schema.Commit):
    check_commit_base(commit_dir, commit)
    parent_id = repo_schema.get_parent_commit_id(name)
    if parent_id is None:
        # initial commit
        assert not (commit_dir / "parent").exists()
        assert not list((commit_dir / "parents").iterdir())
    else:
        parent_commit = repo_schema.get_commit_by_id(parent_id)
        assert (commit_dir / "parent" / "hash").read_text() == parent_commit.hash
        assert [f.name for f in (commit_dir / "parents").iterdir()] == [parent_commit.hash]


def check_merge_commit(repo_schema: schema.Repo, commit_dir: pathlib.Path, name: str, commit: schema.MergeCommit):
    check_commit_base(commit_dir, commit)
    parent_id = repo_schema.get_parent_commit_id(name)
    assert parent_id is not None, "merge commit cannot be initial"
    parent_hashes = [
        repo_schema.get_commit_by_id(parent_id).hash,
        repo_schema.get_commit_by_id(commit.other_commit).hash,
    ]
    assert (commit_dir / "parent" / "hash").read_text() in parent_hashes
    assert sorted(f.name for f in (commit_dir / "parents").iterdir()) == sorted(parent_hashes)


def test_commits(mount: pathlib.Path, repo_schema: schema.Repo):
    hashes = [c.hash for _, c in repo_schema.iter_commits()]
    assert sorted(hashes + ["HEAD"]) == sorted(f.name for f in (mount / "commits").iterdir()), \
        "directories and commits should match"

    for name, commit in repo_schema.iter_commits():
        commit_dir = mount / "commits" / commit.hash
        if isinstance(commit, schema.Commit):
            check_commit(repo_schema, commit_dir, name, commit)
        else:
            check_merge_commit(repo_schema, commit_dir, name, commit)

    check_commit(repo_schema, mount / "commits" / "HEAD", "main:-1", repo_schema.get_commit_by_id("main:-1"))


def test_branches(mount: pathlib.Path, repo_schema: schema.Repo):
    assert sorted(repo_schema.branches) == sorted(f.name for f in (mount / "branches").iterdir()), \
        "directories and branches should match"
