package logging

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_LogLevelFlag(t *testing.T) {
	mapping := map[LogLevelFlag]string{
		Debug:   "DEBUG",
		Info:    "INFO",
		Warning: "WARNING",
		Error:   "ERROR",
	}

	t.Run("flag to string", func(t *testing.T) {
		for k, v := range mapping {
			assert.Equal(t, k.String(), v)
		}
	})

	t.Run("flag from string", func(t *testing.T) {
		for k, v := range mapping {
			var flag LogLevelFlag
			err := flag.Set(v)
			assert.NoError(t, err)
			assert.Equal(t, k, flag)
		}
	})

	t.Run("flag from int", func(t *testing.T) {
		for i := 0; i < len(mapping); i++ {
			var flag LogLevelFlag
			s := strconv.Itoa(i)
			err := flag.Set(s)
			assert.NoError(t, err)
			assert.Equal(t, i, int(flag))
		}
	})

	t.Run("flag from invalid", func(t *testing.T) {
		for _, s := range []string{"-1", strconv.Itoa(len(mapping)), "aaa"} {
			var flag LogLevelFlag
			err := flag.Set(s)
			assert.Error(t, err)
		}
	})

}

func Test_Init(t *testing.T) {
	Init(Warning)
	assert.Equal(t, DebugLog.Writer(), io.Discard)
	assert.Equal(t, InfoLog.Writer(), io.Discard)
	assert.Equal(t, WarningLog.Writer(), os.Stdout)
	assert.Equal(t, ErrorLog.Writer(), os.Stdout)
}

func Test_MakeFileLogger(t *testing.T) {
	tempdir := t.TempDir()
	name := filepath.Join(tempdir, "test.log")
	logger, err := MakeFileLogger(name)
	logger.SetFlags(0)
	assert.Nil(t, err, "Unexpected error when creating file")
	logger.Printf("test")

	file, err := os.Open(name)
	assert.Nil(t, err, "Unexpected error when opening file")
	data := make([]byte, 20)
	n, err := file.Read(data)
	assert.Nil(t, err, "Unexpected error when reading data")
	assert.Equal(t, "test\n", string(data[:n]))
}
