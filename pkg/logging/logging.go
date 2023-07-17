package logging

import (
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
	switch kind {
	case Full:
		name = fullName
	case Class:
		parts := strings.Split(fullName, ".")
		name = strings.Join(parts[len(parts)-2:], ".")
	case Method:
		parts := strings.Split(fullName, ".")
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
