import os
import pathlib
import tempfile

import pytest
import sh

GOGITFS_BINARY = os.environ["GOGITFS_BINARY"]
REPO_URL = "<TODO>"


@pytest.fixture(scope="session")
def repo():
    with tempfile.TemporaryDirectory() as tmpdir:
        sh.git.clone(REPO_URL, tmpdir)
        yield pathlib.Path(tmpdir)
