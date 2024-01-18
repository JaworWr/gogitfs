package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_formatCtxValue(t *testing.T) {
	testCases := []struct {
		valType   string
		valResult string
		val       any
		expected  string
	}{
		{"int", "a number", 5, "5"},
		{"uint8", "a number", uint8(5), "5"},
		{"float", "a floating point number", 5.4, "5.4"},
		{"true", "\"true\"", true, "true"},
		{"false", "\"false\"", false, "false"},
		{"string", "a quoted string with newlines replaced", "abc\ndef", "\"abc;def\""},
	}
	for _, tc := range testCases {
		t.Run(tc.valType, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatCtxValue(tc.val), "%s should be formatted as %s", tc.valType, tc.valResult)
		})
	}
}

func Test_formatCtx(t *testing.T) {
	ctx := CallCtx{
		"a": 5,
		"b": "foo\nbar",
	}
	expected := "a=5, b=\"foo;bar\""
	assert.Equal(t, expected, formatCtx(ctx), "incorrect formatting")
}

func Test_concatCtx(t *testing.T) {
	type args struct {
		dst, src CallCtx
	}

	testCases := []struct {
		name string
		args
		expected CallCtx
	}{
		{
			"nil+ctx",
			args{nil, CallCtx{"a": 5, "b": "foo\nbar"}},
			CallCtx{"a": 5, "b": "foo\nbar"},
		},
		{
			"ctx+nil",
			args{CallCtx{"a": 7, "b": "foo\nbaz"}, nil},
			CallCtx{"a": 7, "b": "foo\nbaz"},
		},
		{
			"ctx+ctx",
			args{CallCtx{"a1": 7, "b1": "foo\nbaz"}, CallCtx{"a": 5, "b": "foo\nbar"}},
			CallCtx{"a1": 7, "b1": "foo\nbaz", "a": 5, "b": "foo\nbar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := concatCtx(tc.dst, tc.src)
			assert.Equal(t, tc.expected, result, "incorrect result context")
		})
	}
}

func funcNames() (full, pkg, class, method string) {
	full = CurrentFuncName(0, Full)
	pkg = CurrentFuncName(0, Package)
	class = CurrentFuncName(0, Class)
	method = CurrentFuncName(0, Method)
	return
}

type sampleClass struct{}

func (c sampleClass) methodNames() (full, pkg, class, method string) {
	full = CurrentFuncName(0, Full)
	pkg = CurrentFuncName(0, Package)
	class = CurrentFuncName(0, Class)
	method = CurrentFuncName(0, Method)
	return
}

func Test_CurrentFuncName(t *testing.T) {
	type expected struct {
		full, pkg, class, method string
	}

	testCases := []struct {
		name string
		f    func() (string, string, string, string)
		expected
	}{
		{
			"function",
			funcNames,
			expected{
				"gogitfs/pkg/logging.funcNames",
				"logging.funcNames",
				"funcNames",
				"funcNames",
			},
		},
		{
			"method",
			sampleClass{}.methodNames,
			expected{
				"gogitfs/pkg/logging.sampleClass.methodNames",
				"logging.sampleClass.methodNames",
				"sampleClass.methodNames",
				"methodNames",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			full, pkg, class, method := tc.f()
			assert.Equal(t, tc.full, full, "incorrect result for format Full")
			assert.Equal(t, tc.pkg, pkg, "incorrect result for format Package")
			assert.Equal(t, tc.class, class, "incorrect result for format Class")
			assert.Equal(t, tc.method, method, "incorrect result for format Method")
		})
	}
}
