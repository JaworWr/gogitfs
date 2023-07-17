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

func CurrentFuncName(skip int, kind FuncName) string {
	pc := make([]uintptr, 16)
	n := runtime.Callers(skip+2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return ProcessFuncName(frame.Function, kind)
}
