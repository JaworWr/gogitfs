import contextlib
import os
import pathlib
import subprocess
from typing import Iterable

import sh

from test.gogitfs_tests.conftest import RepoInfo, GOGITFS_BINARY


@contextlib.contextmanager
def mount_with_flags(repo_path, mount_point, flags: Iterable[str], capture_output: bool = False, unmount: bool = True):
    process = subprocess.run(
        [GOGITFS_BINARY, *flags, str(repo_path), str(mount_point)],
        capture_output=capture_output,
        encoding="utf-8",
    )
    try:
        yield process
    finally:
        if unmount and process.returncode == 0:
            sh.umount(str(mount_point))


def test_help_flag(repo: RepoInfo):
    flags = ["-help"]
    with mount_with_flags(repo.path, "dummy", flags, True, False) as process:
        assert process.returncode == 0
        assert process.stderr.startswith(f"Usage: {GOGITFS_BINARY} <repo-dir> <mount-dir>\n")


def test_uid_gid(repo: RepoInfo, tmp_path: pathlib.Path):
    uid = 1234
    gid = 5678
    flags = ["-uid", str(uid), "-gid", str(gid)]
    mount_point = tmp_path / "mount"
    mount_point.mkdir()

    with mount_with_flags(repo.path, mount_point, flags) as process:
        assert process.returncode == 0
        for f in mount_point.glob("**"):
            stat = f.stat()
            assert stat.st_uid == uid
            assert stat.st_gid == gid


def test_allow_empty(repo: RepoInfo, tmp_path: pathlib.Path):
    mount_point = tmp_path / "mount"
    mount_point.mkdir()
    (mount_point / "foo").mkdir()

    flags = []
    with mount_with_flags(repo.path, mount_point, flags, True) as process:
        assert process.returncode == 1
        assert (mount_point / "foo").exists(), "repo should not be mounted"

    flags = ["-allow-nonempty"]
    with mount_with_flags(repo.path, mount_point, flags) as process:
        assert process.returncode == 0
        assert (mount_point / "commits").exists(), "repo should be mounted"
        assert not (mount_point / "foo").exists(), "repo should be mounted"
