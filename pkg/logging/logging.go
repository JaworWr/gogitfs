package logging

type LogLevelFlag int

const (
	Debug LogLevelFlag = iota
	Info
	Warning
	Error
)

var LogLevel LogLevelFlag = Info
