package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatCtxValue(t *testing.T) {
	assert.Equal(t, "5", formatCtxValue(5), "int should be formatted as a number")
	assert.Equal(t, "5.4", formatCtxValue(5.4), "float should be formatted as a number")
	assert.Equal(t, "true", formatCtxValue(true), "true should be formatted as \"true\"")
	assert.Equal(t, "false", formatCtxValue(false), "false should be formatted as \"false\"")
	assert.Equal(t, "\"abc;def\"", formatCtxValue("abc\ndef"),
		"strings should be quoted and have newlines removed")
}
