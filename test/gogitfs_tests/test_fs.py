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


def test_commits(mount: pathlib.Path, repo_schema: schema.Repo):
    hashes = [c.hash for _, c in repo_schema.iter_commits()]
    assert sorted(hashes + ["HEAD"]) == sorted(f.name for f in (mount / "commits").iterdir()), \
        "directories and commits should match"

    def check_commit_base(dir_: str, commit: schema.Commit | schema.MergeCommit):
        commit_dir = mount / "commits" / dir_
        assert (commit_dir / "hash").read_text() == commit.hash
        assert (commit_dir / "message").read_text() == commit.message
        for f in commit_dir.glob("**"):
            stat = os.stat(f)
            assert stat.st_mtime == stat.st_atime == stat.st_ctime == commit.time.timestamp()

    def check_commit(dir_: str, name: str, commit: schema.Commit):
        check_commit_base(dir_, commit)
        commit_dir = mount / "commits" / dir_
        parent_id = repo_schema.get_parent_commit_id(name)
        if parent_id is None:
            # initial commit
            assert not (commit_dir / "parent").exists()
            assert not list((commit_dir / "parents").iterdir())
        else:
            parent_commit = repo_schema.get_commit_by_id(parent_id)
            assert (commit_dir / "parent" / "hash").read_text() == parent_commit.hash
            assert [f.name for f in (commit_dir / "parents").iterdir()] == [parent_commit.hash]

    def check_merge_commit(dir_: str, name: str, commit: schema.MergeCommit):
        check_commit_base(dir_, commit)
        commit_dir = mount / "commits" / dir_
        parent_id = repo_schema.get_parent_commit_id(name)
        assert parent_id is not None, "merge commit cannot be initial"
        parent_hashes = [
            repo_schema.get_commit_by_id(parent_id).hash,
            repo_schema.get_commit_by_id(commit.other_commit).hash,
        ]
        assert (commit_dir / "parent" / "hash").read_text() in parent_hashes
        assert sorted(f.name for f in (commit_dir / "parents").iterdir()) == sorted(parent_hashes)

    for name, commit in repo_schema.iter_commits():
        if isinstance(commit, schema.Commit):
            check_commit(commit.hash, name, commit)
        else:
            check_merge_commit(commit.hash, name, commit)

    check_commit("HEAD", "main:-1", repo_schema.get_commit_by_id("main:-1"))
