package log

type Level int
type Logger interface {
	SetLevel(lvl Level)
	SetPrefix(prefix string)
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Fatal(a ...interface{})
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Fatalf(format string, a ...interface{})
	Panic(a ...interface{})
	Panicf(format string, a ...interface{})
}
