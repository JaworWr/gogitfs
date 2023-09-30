import contextlib
import pathlib
import subprocess
from typing import Iterable

import sh

from test.gogitfs_tests.conftest import RepoInfo, GOGITFS_BINARY


@contextlib.contextmanager
def mount_with_flags(repo_path, mount_point, flags: Iterable[str]):
    process = subprocess.run(
        [GOGITFS_BINARY, *flags, str(repo_path), str(mount_point)]
    )
    try:
        yield process
    finally:
        if process.returncode == 0:
            sh.umount(str(mount_point))


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
