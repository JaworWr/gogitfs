package mountpoint

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path"
	"testing"
)

func Test_ValidateMountpoint(t *testing.T) {
	// helper function
	tmpdir := t.TempDir()
	tmpPath := func(parts ...string) string {
		parts = append([]string{tmpdir}, parts...)
		return path.Join(parts...)
	}
	mkfile := func(parts ...string) {
		p := tmpPath(parts...)
		f, err := os.Create(p)
		if err != nil {
			t.Fatalf("Cannot create file %v: %v", p, err)
		}
		_ = f.Close()
	}
	mkdir := func(perm os.FileMode, parts ...string) {
		p := tmpPath(parts...)
		err := os.Mkdir(p, perm)
		if err != nil {
			t.Fatalf("Cannot create directory %v: %v", p, err)
		}
	}

	// setup files for tests
	mkfile("file")
	mkdir(0555, "readonly")
	mkdir(0777|fs.ModeSticky, "sticky")
	mkdir(0777, "empty")
	mkdir(0777, "nonempty")
	mkfile("nonempty", "foo")

	// run tests
	type args struct {
		filename      string
		allowNonEmpty bool
	}
	testCases := []struct {
		name string
		args
		shouldWork bool
	}{
		{"absent", args{"absent", true}, false},
		{"file", args{"file", true}, false},
		{"readonly", args{"readonly", true}, false},
		{"sticky", args{"file", true}, false},
		{"nonempty allowed", args{"nonempty", true}, true},
		{"nonempty not allowed", args{"nonempty", false}, false},
		{"empty allowed", args{"empty", true}, true},
		{"empty not allowed", args{"empty", false}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := tmpPath(tc.filename)
			_, err := ValidateMountpoint(p, tc.allowNonEmpty)
			if tc.shouldWork {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
