# gogitfs

Mount a git repository as a FUSE filesystem.

## Usage
To simply mount a repository at a given directory, run: 
```shell
gogitfs <repository-path> <mount-path>
```

For a full list of options run:
```shell
gogitfs -h
```

## Directory structure
The repository is presented as a directory containing two subdirectories:
* `commits` - contains a single directory per commit, and a symlink to the head commit called simply `HEAD`
* `branches` - contains a single directory per branch, each containing commits on that branch.

The directory of each commit has the following structure:
```text
├── hash
├── log
├── message
├── parent -> <parent commit>
└── parents
```
`hash` and `message` are text files containing, respectively, the commit hash string and the message text. 
The directory `parents` contains symlinks to all parent commits. If the commit has parents, a symlink 
called `parent` will be created pointing to the first parent. The directory `log` contains the git log starting
at the current commit (but not including it).