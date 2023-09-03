# mountpoint

Package mountpoint provides utilities for mountpoint validation.

We assume that a valid mountpoint should have the following features:
* it is a directory
* current user has write access on it
* sticky bit is not set
* the directory is empty. This check can be disabled.

For more information see https://www.kernel.org/doc/html/next/filesystems/fuse.html#how-do-non-privileged-mounts-work.
