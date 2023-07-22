package logging

import (
	"fmt"
	"runtime"
	"strings"
)

type FuncName int

const (
	// Full fully qualified name
	Full FuncName = iota
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
	case Class:
		name = strings.Join(parts[len(parts)-2:], ".")
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

type CallCtx = map[string]string

type CallCtxGetter interface {
	GetCallCtx() CallCtx
}

func formatCtx(ctx CallCtx) string {
	parts := make([]string, 0)
	for k, v := range ctx {
		parts = append(parts, fmt.Sprintf("%v=\"%v\"", k, v))
	}
	return strings.Join(parts, ", ")
}

func concatMaps(dst map[string]string, src map[string]string) map[string]string {
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
	info := l.GetCallCtx()
	info = concatMaps(info, extra)
	methodInfo := formatCtx(info)
	DebugLog.Printf("Called %v (%v)", methodName, methodInfo)
}

func BoolToStr(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}
