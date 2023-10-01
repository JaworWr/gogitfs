import contextlib
import pathlib
import subprocess
from typing import Iterable

import sh

from test.gogitfs_tests.conftest import GOGITFS_BINARY


@contextlib.contextmanager
def mount_with_flags(repo_path, mount_point, flags: Iterable[str], capture_output: bool = False, unmount: bool = True):
    args = [GOGITFS_BINARY, *flags]
    if repo_path is not None:
        args.append(str(repo_path))
    if mount_point is not None:
        args.append(str(mount_point))
    else:
        unmount = False
    process = subprocess.run(
        args,
        capture_output=capture_output,
        encoding="utf-8",
    )
    try:
        yield process
    finally:
        if unmount and process.returncode == 0:
            sh.umount(str(mount_point))


def is_usage_line(line: str) -> bool:
    return line.strip() == f"Usage: {GOGITFS_BINARY} <repo-dir> <mount-dir>"


def is_filesystem_error(msg: str, err: str) -> bool:
    err_msg = f"cannot start the filesystem daemon\n{err}"
    return msg.startswith(err_msg)


def test_help_flag():
    for flags in [["-help"], ["-h"]]:
        with mount_with_flags(None, None, flags, True, False) as process:
            assert process.returncode == 0, f"error for flags {flags}"
            first_line = process.stderr.splitlines()[0]
            assert is_usage_line(first_line), f"invalid first line for flags {flags}"


def test_invalid_args(repo_path: pathlib.Path, tmp_path: pathlib.Path):
    examples = [
        ("both missing", None, None, [], "not enough positional arguments"),
        ("second missing", repo_path, None, [], "not enough positional arguments"),
        ("too many args", repo_path, tmp_path, ["foo"], "unexpected arguments"),
        ("invalid flag", repo_path, tmp_path, ["-foo"], "flag provided but not defined"),
        ("invalid arg", repo_path, tmp_path, ["-uid", "aaa"], "invalid value"),
        ("missing arg", None, None, ["-uid"], "flag needs an argument")
    ]

    for name, path1, path2, flags, err in examples:
        with mount_with_flags(path1, path2, flags, True) as process:
            assert process.returncode != 0, f"process didn't fail for case {name}"
            assert process.stderr.startswith(err), f"wrong error message for case {name}"
            try:
                second_line = process.stderr.splitlines()[1]
            except IndexError:
                second_line = "TOO SHORT"
            assert is_usage_line(second_line), f"help wasn't shown for case {name}"


def test_uid_gid(repo_path: pathlib.Path, tmp_path: pathlib.Path):
    uid = 1234
    gid = 5678
    flags = ["-uid", str(uid), "-gid", str(gid)]
    mount_point = tmp_path / "mount"
    mount_point.mkdir()

    with mount_with_flags(repo_path, mount_point, flags) as process:
        assert process.returncode == 0
        for f in mount_point.glob("**"):
            stat = f.stat()
            assert stat.st_uid == uid
            assert stat.st_gid == gid


def test_allow_empty(repo_path: pathlib.Path, tmp_path: pathlib.Path):
    mount_point = tmp_path / "mount"
    mount_point.mkdir()
    (mount_point / "foo").mkdir()

    flags = []
    with mount_with_flags(repo_path, mount_point, flags, True) as process:
        assert process.returncode == 1
        assert is_filesystem_error(process.stderr, "directory not empty")
        assert (mount_point / "foo").exists(), "repo should not be mounted"

    flags = ["-allow-nonempty"]
    with mount_with_flags(repo_path, mount_point, flags) as process:
        assert process.returncode == 0
        assert (mount_point / "commits").exists(), "repo should be mounted"
        assert not (mount_point / "foo").exists(), "repo should be mounted"
