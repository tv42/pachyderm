package fuse

import (
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
)

// Exported for test use only.

func NewFilesystem(
	apiClient pfs.APIClient,
	shard *pfs.Shard,
	commitMounts []*CommitMount,
) fs.FS {
	return newFilesystem(apiClient, shard, commitMounts)
}
