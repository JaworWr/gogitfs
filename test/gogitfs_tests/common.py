import contextlib
import os
import re
import subprocess
from typing import Iterable

import sh

GOGITFS_BINARY = os.environ["GOGITFS_BINARY"]


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


def is_filesystem_error(msg: str, err_pattern: str) -> bool:
    lines = msg.splitlines()
    if len(lines) < 2:
        return False
    if lines[0].strip() != "cannot start the filesystem daemon":
        return False
    return bool(re.match(err_pattern, lines[1]))
