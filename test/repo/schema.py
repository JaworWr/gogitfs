import datetime as dt
from dataclasses import dataclass
from typing import Iterable

import dataclasses_json
from dataclasses_json import dataclass_json

dataclasses_json.cfg.global_config.encoders[dt.datetime] = dt.datetime.isoformat
dataclasses_json.cfg.global_config.decoders[dt.datetime] = dt.datetime.fromisoformat


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
    hash: str | None = None


@dataclass_json
@dataclass
class MergeCommit:
    message: str
    time: dt.datetime
    other_commit: str
    hash: str | None = None


@dataclass_json
@dataclass
class Branch:
    from_commit: str | None
    commits: list[Commit | MergeCommit]


@dataclass_json
@dataclass
class Repo:
    branches: dict[str, Branch]
    main_branch: str
    active_branch: str

    def iter_commits(self) -> Iterable[tuple[str, Commit | MergeCommit]]:
        for name, branch in self.branches.items():
            for i, commit in enumerate(branch.commits):
                id_ = f"{name}:{i}"
                yield id_, commit

    def get_commit_by_id(self, id_: str) -> Commit | MergeCommit:
        branch, idx = id_.split(":")
        idx = int(idx)
        return self.branches[branch].commits[idx]

    def get_parent_commit_id(self, id_: str) -> str | None:
        branch, idx = id_.split(":")
        idx = int(idx)
        if idx != 0:
            parent_id = f"{branch}:{idx - 1}"
        else:
            parent_id = self.branches[branch].from_commit
        return parent_id
