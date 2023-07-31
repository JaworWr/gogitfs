package logging

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type FuncName int

const (
	// Full fully qualified name
	Full FuncName = iota
	// Package package + class + method name
	Package
	// Class class + method name
	Class
	// Method method name only
	Method
)

// ProcessFuncName extract desired parts from fully qualified function name
func ProcessFuncName(fullName string, kind FuncName) (name string) {
	pathParts := strings.Split(fullName, "/")
	parts := strings.Split(pathParts[len(pathParts)-1], ".")
	switch kind {
	case Full:
		name = fullName
	case Package:
		name = pathParts[len(pathParts)-1]
	case Class:
		if len(parts) <= 2 {
			name = parts[len(parts)-1]
		} else {
			name = strings.Join(parts[len(parts)-2:], ".")
		}
	case Method:
		name = parts[len(parts)-1]
	}
	return
}

// CurrentFuncName get current function name
// skip: how many frames to skip after the caller of CurrentFuncName
// i.e. pass 0 to call the coller of CurrentFuncName, 1 for its caller etc.
// kind: which part of the function name to return
func CurrentFuncName(skip int, kind FuncName) string {
	pc := make([]uintptr, 16)
	n := runtime.Callers(skip+2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return ProcessFuncName(frame.Function, kind)
}

type CallCtx = map[string]any

type CallCtxGetter interface {
	GetCallCtx() CallCtx
}

func formatCtxValue(v any) string {
	useQuotes := true
	var formatted string
	switch v.(type) {
	case int, int8, int16, int32, int64:
		useQuotes = false
	case uint, uint8, uint16, uint32, uint64, uintptr:
		useQuotes = false
	case float32, float64, complex64, complex128:
		useQuotes = false
	case bool:
		if v.(bool) {
			formatted = "true"
		} else {
			formatted = "false"
		}
	}
	if formatted == "" {
		if useQuotes {
			formatted = fmt.Sprintf("\"%v\"", v)
		} else {
			formatted = fmt.Sprintf("%v", v)
		}
	}
	formatted = strings.Replace(formatted, "\n", ";", -1)
	return formatted
}

func formatCtx(ctx CallCtx) string {
	var keys []string
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		v := formatCtxValue(ctx[k])
		parts = append(parts, fmt.Sprintf("%v=%v", k, v))
	}
	return strings.Join(parts, ", ")
}

func concatCtx(dst CallCtx, src CallCtx) CallCtx {
	if dst == nil {
		return src
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// LogCall log function call
// the format is Called <method> (<key>=<value>)
// with key, value returned by GetCallCtx()
func LogCall(l CallCtxGetter, extra CallCtx) {
	methodName := CurrentFuncName(1, Class)
	var info CallCtx
	if l != nil {
		info = l.GetCallCtx()
	}
	info = concatCtx(info, extra)
	methodInfo := formatCtx(info)
	DebugLog.Printf("Called %v (%v)", methodName, methodInfo)
}
