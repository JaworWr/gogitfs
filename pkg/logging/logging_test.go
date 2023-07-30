package logging

import (
	"github.com/stretchr/testify/assert"
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
	for k, v := range mapping {
		assert.Equal(t, k.String(), v, "Incorrect flag string")
	}

	for k, v := range mapping {
		var flag LogLevelFlag
		err := flag.Set(v)
		assert.Nil(t, err, "Unexpected error")
		assert.Equal(t, k, flag)
	}

	for i := 0; i < len(mapping); i++ {
		var flag LogLevelFlag
		s := strconv.Itoa(i)
		err := flag.Set(s)
		assert.Nil(t, err, "Unexpected error")
		assert.Equal(t, i, int(flag))
	}

	for _, s := range []string{"-1", strconv.Itoa(len(mapping)), "aaa"} {
		var flag LogLevelFlag
		err := flag.Set(s)
		assert.NotNil(t, err, "Expected an error")
	}
}
