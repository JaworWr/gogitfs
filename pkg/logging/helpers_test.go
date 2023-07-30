package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_formatCtxValue(t *testing.T) {
	assert.Equal(t, "5", formatCtxValue(5), "int should be formatted as a number")
	assert.Equal(t, "5", formatCtxValue(uint8(5)), "uint8 should be formatted as a number")
	assert.Equal(t, "5.4", formatCtxValue(5.4), "float should be formatted as a number")
	assert.Equal(t, "true", formatCtxValue(true), "true should be formatted as \"true\"")
	assert.Equal(t, "false", formatCtxValue(false), "false should be formatted as \"false\"")
	assert.Equal(t, "\"abc;def\"", formatCtxValue("abc\ndef"),
		"strings should be quoted and have newlines removed")
}

func Test_formatCtx(t *testing.T) {
	ctx := CallCtx{
		"a": 5,
		"b": "foo\nbar",
	}
	expected := "a=5, b=\"foo;bar\""
	assert.Equal(t, expected, formatCtx(ctx))
}

func Test_concatCtx(t *testing.T) {
	var ctx1, ctx2 CallCtx

	ctx1 = nil
	ctx2 = CallCtx{
		"a": 5,
		"b": "foo\nbar",
	}
	assert.Equal(t, ctx2, concatCtx(ctx1, ctx2))

	ctx1 = CallCtx{
		"a": 5,
		"b": "foo\nbar",
	}
	ctx2 = nil
	assert.Equal(t, ctx1, concatCtx(ctx1, ctx2))

	ctx1 = CallCtx{
		"a": 5,
		"b": "foo\nbar",
	}
	ctx2 = CallCtx{
		"a1": 5,
		"b1": "baz",
	}
	expected := CallCtx{
		"a":  5,
		"b":  "foo\nbar",
		"a1": 5,
		"b1": "baz",
	}
	assert.Equal(t, expected, concatCtx(ctx1, ctx2))
}

func funcNames() (full string, pkg string, class string, method string) {
	full = CurrentFuncName(0, Full)
	pkg = CurrentFuncName(0, Package)
	class = CurrentFuncName(0, Class)
	method = CurrentFuncName(0, Method)
	return
}

type sampleClass struct{}

func (c sampleClass) methodNames() (full string, pkg string, class string, method string) {
	full = CurrentFuncName(0, Full)
	pkg = CurrentFuncName(0, Package)
	class = CurrentFuncName(0, Class)
	method = CurrentFuncName(0, Method)
	return
}

func Test_CurrentFuncName(t *testing.T) {
	var full, pkg, class, method string
	full, pkg, class, method = funcNames()
	assert.Equal(t, "gogitfs/pkg/logging.funcNames", full)
	assert.Equal(t, "logging.funcNames", pkg)
	assert.Equal(t, "funcNames", class)
	assert.Equal(t, "funcNames", method)

	full, pkg, class, method = sampleClass{}.methodNames()
	assert.Equal(t, "gogitfs/pkg/logging.sampleClass.methodNames", full)
	assert.Equal(t, "logging.sampleClass.methodNames", pkg)
	assert.Equal(t, "sampleClass.methodNames", class)
	assert.Equal(t, "methodNames", method)
}
