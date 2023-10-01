import pathlib

from test.gogitfs_tests.common import mount_with_flags, is_usage_line


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
