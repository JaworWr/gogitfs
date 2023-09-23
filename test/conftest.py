import os

import pytest
import sh

GOGITFS_BINARY = os.environ["GOGITFS_BINARY"]
REPO_URL = "<TODO>"


@pytest.fixture(scope="session")
def repo(tmp_path):
    repo_path = tmp_path / "repo"
    sh.git.clone(REPO_URL, str(repo_path))
    return repo_path
