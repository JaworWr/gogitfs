from test.gogitfs_tests.conftest import RepoInfo


def test_dummy(repo: RepoInfo):
    assert repo.path.is_dir()
