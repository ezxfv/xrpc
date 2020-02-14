package log

import "github.com/prometheus/client_golang/prometheus"

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

	EnableCounter(labelNames ...string) *prometheus.CounterVec
}

// SetLevel sets the log level of ALL receivers
func SetLevel(lvl Level) {
	gLogger.SetLevel(lvl)
}

// SetPrefix sets the prefix of ALL receivers
func SetPrefix(prefix string) {
	gLogger.SetPrefix(prefix)
}

// Debug logs arguments
func Debug(a ...interface{}) {
	gLogger.Debug(a...)
}

// Info logs arguments
func Info(a ...interface{}) {
	gLogger.Info(a...)
}

// Warn logs arguments
func Warn(a ...interface{}) {
	gLogger.Warn(a...)
}

// Error logs arguments
func Error(a ...interface{}) {
	gLogger.Error(a...)
}

// Fatal logs arguments
func Fatal(a ...interface{}) {
	gLogger.Fatal(a...)
}

// Panic logs arguments
func Panic(a ...interface{}) {
	gLogger.Panic(a...)
}

// Debugf logs formated arguments
func Debugf(format string, a ...interface{}) {
	gLogger.Debugf(format, a...)
}

// Infof logs formated arguments
func Infof(format string, a ...interface{}) {
	gLogger.Infof(format, a...)
}

// Warnf logs formated arguments
func Warnf(format string, a ...interface{}) {
	gLogger.Warnf(format, a...)
}

// Errorf logs formated arguments
func Errorf(format string, a ...interface{}) {
	gLogger.Errorf(format, a...)
}

// Fatalf logs formated arguments
func Fatalf(format string, a ...interface{}) {
	gLogger.Fatalf(format, a...)
}

// Panicf logs formated arguments
func Panicf(format string, a ...interface{}) {
	gLogger.Panicf(format, a...)
}
