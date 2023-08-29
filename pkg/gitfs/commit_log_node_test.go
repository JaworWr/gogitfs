package gitfs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_commitSymlink(t *testing.T) {
	repo, extras := makeRepo(t)
	basePath := "asdf"
	type args struct {
		commit   string
		basePath *string
	}
	testCases := []struct {
		name string
		args
		expectedPrefix string
	}{
		{
			"foo without prefix",
			args{"foo", nil},
			"",
		},
		{
			"bar with prefix",
			args{"bar", &basePath},
			"asdf/",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := extras.commits[tc.commit]
			commit, err := repo.CommitObject(hash)
			if err != nil {
				t.Fatalf("Error during retrieval of commit object: %v", err)
			}
			node := commitSymlink(commit, tc.basePath)

			path := string(node.Data)
			assert.Equal(t, tc.expectedPrefix+hash.String(), path, "incorrect path")

			commitTime := uint64(commitSignatures[tc.commit].When.Unix())
			assert.Equal(t, commitTime, node.Attr.Mtime, "incorrect mtime")
			assert.Equal(t, commitTime, node.Attr.Atime, "incorrect atime")
			assert.Equal(t, commitTime, node.Attr.Ctime, "incorrect ctime")
		})
	}
}
