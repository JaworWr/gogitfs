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
			assert.Equal(t, k.String(), v, "incorrect string from flag")
		}
	})

	t.Run("flag from string", func(t *testing.T) {
		for k, v := range mapping {
			var flag LogLevelFlag
			err := flag.Set(v)
			assert.NoError(t, err, "unexpected error during flag conversion")
			assert.Equal(t, k, flag, "incorrect flag from string")
		}
	})

	t.Run("flag from int", func(t *testing.T) {
		for i := 0; i < len(mapping); i++ {
			var flag LogLevelFlag
			s := strconv.Itoa(i)
			err := flag.Set(s)
			assert.NoError(t, err, "unexpected error during flag cconversion")
			assert.Equal(t, i, int(flag), "incorrect flag from int")
		}
	})

	t.Run("flag from invalid", func(t *testing.T) {
		for _, s := range []string{"-1", strconv.Itoa(len(mapping)), "aaa"} {
			var flag LogLevelFlag
			err := flag.Set(s)
			assert.Error(t, err, "should get an error for invalid flag")
		}
	})

}

func Test_Init(t *testing.T) {
	Init(Warning)
	assert.Equal(t, DebugLog.Writer(), io.Discard, "incorrect IO for DEBUG")
	assert.Equal(t, InfoLog.Writer(), io.Discard, "incorrect IO for INFO")
	assert.Equal(t, WarningLog.Writer(), os.Stdout, "incorrect IO for WARNING")
	assert.Equal(t, ErrorLog.Writer(), os.Stdout, "incorrect IO for ERROR")
}

func Test_MakeFileLogger(t *testing.T) {
	tempdir := t.TempDir()
	name := filepath.Join(tempdir, "test.log")
	logger, err := MakeFileLogger(name)
	logger.SetFlags(0)
	assert.NoError(t, err, "unexpected file creation error")
	logger.Printf("test")

	file, err := os.Open(name)
	assert.NoError(t, err, "unexpected file opening error")
	data := make([]byte, 20)
	n, err := file.Read(data)
	assert.NoError(t, err, "unexpected data reading error")
	assert.Equal(t, "test\n", string(data[:n]), "read incorrect data")
}
