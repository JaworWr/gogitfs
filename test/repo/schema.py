from dataclasses import dataclass
import datetime as dt

import dataclasses_json
from dataclasses_json import dataclass_json


dataclasses_json.cfg.global_config.encoders[dt.datetime] = dt.datetime.isoformat
dataclasses_json.cfg.global_config.decoders[dt.datetime] = dt.datetime.fromisoformat


__all__ = ["CommitFile", "Commit", "Branch", "Repo"]


NULL_HASH = "<NIL>"


@dataclass_json
@dataclass
class CommitFile:
    path: str
    contents: str


@dataclass_json
@dataclass
class Commit:
    message: str
    time: dt.datetime
    files: list[CommitFile]
    hash: str = NULL_HASH


@dataclass_json
@dataclass
class Branch:
    from_commit: str | None
    commits: list[Commit]


@dataclass_json
@dataclass
class Repo:
    branches: dict[str, Branch]
    main_branch: str
