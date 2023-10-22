import os
import pathlib

from test.gogitfs_tests.common import mount_with_flags, is_usage_line, is_filesystem_error


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


def test_nonexistent_repo(tmp_path: pathlib.Path):
    with mount_with_flags(tmp_path / "repo", tmp_path, [], True) as process:
        assert process.returncode != 0
        assert is_filesystem_error(process.stderr, "cannot create root node: repository does not exist")


def test_invalid_repo(tmp_path: pathlib.Path):
    repo_path = tmp_path / "repo"
    repo_path.mkdir()
    mount_point = tmp_path / "mount"
    mount_point.mkdir()
    with mount_with_flags(repo_path, mount_point, [], True) as process:
        assert process.returncode != 0
        assert is_filesystem_error(process.stderr, "cannot create root node: repository does not exist")


def test_invalid_mountpoint(repo_path: pathlib.Path, tmp_path: pathlib.Path):
    mount_point = tmp_path / "mount"
    with mount_with_flags(repo_path, mount_point, [], True) as process:
        assert process.returncode != 0
        assert is_filesystem_error(process.stderr, r"invalid mountpoint: stat .*: no such file or directory")

    mount_point.mkdir()
    os.chmod(mount_point, 0o444)  # make mount point read-only
    with mount_with_flags(repo_path, mount_point, [], True) as process:
        assert process.returncode != 0
        assert is_filesystem_error(process.stderr, "invalid mountpoint: permission denied")
