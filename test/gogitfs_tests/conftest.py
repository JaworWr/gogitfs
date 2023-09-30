import os
import pathlib
import tempfile

import pytest

GOGITFS_BINARY = os.environ["GOGITFS_BINARY"]


@pytest.fixture(scope="session")
def repo():
    with tempfile.TemporaryDirectory() as tmpdir:
        yield pathlib.Path(tmpdir)
