# inode_manager

Package inode_manager contains types meant for storage and lazy creation of Inodes.

This package defines the following types:

* `AttrStore` - generates and stores file attributes (inode number and generation)
* `InodeStore`- generates and stores actual Inodes
* `InodeCache`- combines both to provide a fully managed node cache. Individual components may be accessed if necessary

All types provide a method `GetOrInsert`, which performs the lazy creation of the object, retrieving existing
object if available.
