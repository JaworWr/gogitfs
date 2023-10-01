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
    hashes = [c.hash for c in repo_schema.iter_commits()]
    assert sorted(hashes + ["HEAD"]) == sorted(f.name for f in (mount / "commits").iterdir()), \
        "directories and commits should match"


