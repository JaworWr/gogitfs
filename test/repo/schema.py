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
    """File added in a commit, represented as its path and contents.

    The path must be relative to the repository root.
    """
    path: str
    contents: str


@dataclass_json
@dataclass
class Commit:
    """A non-merge commit, adding or changing an arbitrary number of files."""
    message: str
    time: dt.datetime
    files: list[CommitFile]
    hash: str | None = None


@dataclass_json
@dataclass
class MergeCommit:
    """A merge commit."""
    message: str
    time: dt.datetime
    other_commit: str
    hash: str | None = None


@dataclass_json
@dataclass
class Branch:
    """A branch starting at a specified commit and containing some new commits."""
    from_commit: str | None
    commits: list[Commit | MergeCommit]


@dataclass_json
@dataclass
class Repo:
    """Repository, specified as a list of branches.

    main_branch should contain the initial commit. active_branch is the current checked out branch.
    """
    branches: dict[str, Branch]
    main_branch: str
    active_branch: str

    def iter_commits(self) -> Iterable[tuple[str, Commit | MergeCommit]]:
        """Iterate over all commits, yielding each commit exactly once"""
        for name, branch in self.branches.items():
            for i, commit in enumerate(branch.commits):
                id_ = f"{name}:{i}"
                yield id_, commit

    def iter_branch_commits(self, branch: str, up_to: int | None = None) -> Iterable[tuple[str, Commit | MergeCommit]]:
        """Iterate over all commits reachable from a branch

        Note that this method makes no attempt to deduplicate, result size may be exponential.
        """
        commits = self.branches[branch].commits
        if up_to is not None:
            commits = commits[:up_to + 1]
        commits = list(enumerate(commits))
        for i, commit in commits[::-1]:
            yield f"{branch}:{i}", commit
            if isinstance(commit, MergeCommit):
                merged_branch, idx = commit.other_commit.split(":")
                yield from self.iter_branch_commits(merged_branch, int(idx))
        if self.branches[branch].from_commit is not None:
            next_branch, idx = self.branches[branch].from_commit.split(":")
            yield from self.iter_branch_commits(next_branch, int(idx))

    def get_commit_by_id(self, commit_id: str) -> Commit | MergeCommit:
        branch, idx = commit_id.split(":")
        idx = int(idx)
        return self.branches[branch].commits[idx]

    def get_parent_commit_id(self, commit_id: str) -> str | None:
        branch, idx = commit_id.split(":")
        idx = int(idx)
        if idx != 0:
            parent_id = f"{branch}:{idx - 1}"
        else:
            parent_id = self.branches[branch].from_commit
        return parent_id
