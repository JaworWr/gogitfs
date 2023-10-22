import pathlib

from test.gogitfs_tests.common import mount_with_flags, is_usage_line, is_filesystem_error


def test_help_flag():
    for flags in [["-help"], ["-h"]]:
        with mount_with_flags(None, None, flags, True, False) as process:
            assert process.returncode == 0, f"error for flags {flags}"
            first_line = process.stderr.splitlines()[0]
            assert is_usage_line(first_line), f"invalid first line for flags {flags}"


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
        assert is_filesystem_error(process.stderr, f"invalid mountpoint: {mount_point}: directory not empty")
        assert (mount_point / "foo").exists(), "repo should not be mounted"

    flags = ["-allow-nonempty"]
    with mount_with_flags(repo_path, mount_point, flags) as process:
        assert process.returncode == 0
        assert (mount_point / "commits").exists(), "repo should be mounted"
        assert not (mount_point / "foo").exists(), "repo should be mounted"
