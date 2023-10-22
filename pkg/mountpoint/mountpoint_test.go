package mountpoint

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path"
	"syscall"
	"testing"
)

func Test_ValidateMountpoint(t *testing.T) {
	// helper functions
	tmpdir := t.TempDir()
	tmpPath := func(parts ...string) string {
		parts = append([]string{tmpdir}, parts...)
		return path.Join(parts...)
	}
	mkFile := func(parts ...string) {
		p := tmpPath(parts...)
		f, err := os.Create(p)
		if err != nil {
			t.Fatalf("Cannot create file %v: %v", p, err)
		}
		_ = f.Close()
	}
	mkDir := func(perm os.FileMode, parts ...string) {
		p := tmpPath(parts...)
		err := os.Mkdir(p, perm)
		if err != nil {
			t.Fatalf("Cannot create directory %v: %v", p, err)
		}
	}

	// setup files for tests
	mkFile("file")
	mkDir(0555, "readonly")
	mkDir(0777|fs.ModeSticky, "sticky")
	mkDir(0777, "empty")
	mkDir(0777, "nonempty")
	mkFile("nonempty", "foo")

	// run tests
	type args struct {
		filename      string
		allowNonEmpty bool
	}
	testCases := []struct {
		name string
		args
		errorType     any
		errorInstance error
	}{
		{"absent", args{"absent", true}, fs.PathError{}, nil},
		{"file", args{"file", true}, nil, ErrNotADirectory},
		{"readonly", args{"readonly", true}, syscall.Errno(0), nil},
		{"sticky", args{"sticky", true}, nil, ErrStickyBitSet},
		{"nonempty allowed", args{"nonempty", true}, nil, nil},
		{"nonempty not allowed", args{"nonempty", false}, nil, ErrNotEmpty},
		{"empty allowed", args{"empty", true}, nil, nil},
		{"empty not allowed", args{"empty", false}, nil, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := tmpPath(tc.filename)
			_, err := ValidateMountpoint(p, tc.allowNonEmpty)
			if tc.errorInstance != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.errorInstance), "incorrect error type")
			} else if tc.errorType != nil {
				assert.Error(t, err)
				assert.True(t, errors.As(err, &tc.errorType), "incorrect error type")
			} else {
				assert.NoError(t, err, "unexpected error returned by ValidateError")
			}
		})
	}
}
