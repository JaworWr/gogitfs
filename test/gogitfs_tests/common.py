import contextlib
import os
import re
import subprocess
from collections.abc import Generator
from typing import Iterable

import sh

GOGITFS_BINARY = os.environ["GOGITFS_BINARY"]


@contextlib.contextmanager
def mount_with_flags(
        repo_path: os.PathLike[str] | None,
        mount_point: os.PathLike[str] | None,
        flags: Iterable[str],
        capture_output: bool = False,
        unmount: bool = True
) -> Generator[subprocess.CompletedProcess[str], None, None]:
    """Run gogitfs binary to mount a given repo at a given mountpoint.

    Args:
        repo_path: repository path to pass to the gogitfs binary. If None, no value will be passed.
        mount_point: mount path to pass to the gogitfs binary. If None, no value will be passed.
        flags: CLI flags to pass to the gogitfs binary
        capture_output: if True, stdout and stderr will be captured
        unmount: whether to call `umount` at mount_point at exit

    Returns:
        a generator yielding a single subprocess.CompletedProcess[str]
    """
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
    """Check if the line is a valid usage description line."""
    return line.strip() == f"Usage: {GOGITFS_BINARY} <repo-dir> <mount-dir>"


def is_filesystem_error(msg: str, err_pattern: str) -> bool:
    """Check if an error matches the expected pattern of a filesystem error."""
    lines = msg.splitlines()
    if len(lines) < 2:
        return False
    if lines[0].strip() != "cannot start the filesystem daemon":
        return False
    return bool(re.match(err_pattern, lines[1]))
