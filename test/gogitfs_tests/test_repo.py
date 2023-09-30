import pathlib
import subprocess

import pytest
import sh

from test.gogitfs_tests.conftest import GOGITFS_BINARY


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
