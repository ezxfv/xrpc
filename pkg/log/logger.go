package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Available log levels
const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Log options for all levels
var (
	optDebug           = &levelOptions{"DEBUG", DEBUG, 34}
	optInfo            = &levelOptions{"INFO ", INFO, 32}
	optWarn            = &levelOptions{"WARN ", WARN, 33}
	optError           = &levelOptions{"ERROR", ERROR, 31}
	optFatal           = &levelOptions{"FATAL", FATAL, 31}
	gLogger            Logger
	DefaultCallerDepth = 3
)

func init() {
	gLogger = NewSimpleDefaultLogger(os.Stdout, DEBUG, "gLogger", true)
}

func SetGlobalLogger(logger Logger) {
	gLogger = logger
}

func GLogger() Logger {
	return gLogger
}

// CallerInfo returns the caller info, default depth is 3,
// caller->logger.Info->logger.logAll->CallerInfo
func CallerInfo() string {
	stack := make([]byte, 1024, 1024)
	n := runtime.Stack(stack, false)
	info := ""
	if n <= 1024 {
		info = string(stack)
	} else {
		info = string(stack[:n])
	}
	callerStack := strings.Split(info, "\n")[2*DefaultCallerDepth+1 : 2*DefaultCallerDepth+3]
	funcInfo := callerStack[0]
	fileInfo := callerStack[1]
	file := fileInfo[strings.LastIndex(fileInfo, "/")+1:]
	return fmt.Sprintf("[%s <%s>]", funcInfo[:strings.LastIndex(funcInfo, "(")], file[:strings.Index(file, " ")])
}

/* Format: [/home/edenzhong/github/ifly/peer.go:120->(*IPeer).Cleanup]
func CallerInfo(level ...int) string {
	l := 3
	if len(level) > 0 && level[0] >= 0 {
		l = level[0]
	}
	_, file, line, ok := runtime.Caller(l)
	if !ok {
		return ""
	}
	//f := runtime.FuncForPC(pc)
	stack := make([]byte, 1024, 1024)
	n := runtime.Stack(stack, false)
	info := ""
	if n <= 1024 {
		info = string(stack)
	} else {
		info = string(stack[:n])
	}
	//funcName := f.name()[strings.LastIndex(f.name(), ".")+1:]
	funcName := ""
	callerStack := strings.Split(info, "\n")[7]
	left := strings.Index(callerStack, "(")
	right := strings.LastIndex(callerStack, "(")
	if left == right {
		funcName = callerStack[:right]
		if mi := strings.Index(funcName, "main."); mi != -1 {
			funcName = funcName[mi+5:]
		} else {
			li := strings.LastIndex(funcName, "/")
			funcName = funcName[li+1:]
			funcName = funcName[strings.Index(funcName, ".")+1:]
		}
	} else {
		funcName = callerStack[left:right]
	}
	return fmt.Sprintf("[%s:%d->%s]", file, line, W*/
// Options to store the key, level and color code of a log
type levelOptions struct {
	Key   string
	Level Level
	Color int
}

// Itol converts an integer to a logo.Level
func Itol(level int) Level {
	switch level {
	case 0:
		return DEBUG
	case 1:
		return INFO
	case 2:
		return WARN
	case 3:
		return ERROR
	case 4:
		return FATAL
	default:
		return DEBUG
	}
}

// DefaultLogger holds all Receivers
type DefaultLogger struct {
	Receivers []*Receiver
	Active    bool
	prefix    string

	enableCounter bool
	counter       *prometheus.CounterVec
}

// NewDefaultLogger returns a new DefaultLogger filled with given Receivers
func NewDefaultLogger(recs ...*Receiver) *DefaultLogger {
	l := &DefaultLogger{
		Active:    true, // Every gLogger is active by default
		Receivers: recs,
	}
	return l
}

// NewSimpleDefaultLogger returns a gLogger with one simple Receiver
func NewSimpleDefaultLogger(w io.Writer, lvl Level, prefix string, color bool) *DefaultLogger {
	l := &DefaultLogger{}
	r := NewReceiver(w, prefix)
	r.Color = color
	r.Level = lvl
	l.Receivers = []*Receiver{r}
	l.Active = true
	return l
}

// SetLevel sets the log level of ALL receivers
func (l *DefaultLogger) SetLevel(lvl Level) {
	for _, r := range l.Receivers {
		r.Level = lvl
	}
}

// SetPrefix sets the prefix of ALL receivers
func (l *DefaultLogger) SetPrefix(prefix string) {
	for _, r := range l.Receivers {
		r.SetPrefix(prefix)
	}
}

// Write to all Receivers
func (l *DefaultLogger) logAll(opt *levelOptions, s string) {
	// Skip everything if gLogger is disabled
	if !l.Active {
		return
	}
	if l.enableCounter {
		l.counter.WithLabelValues(opt.Key).Inc()
	}
	callerInfo := ""
	//if !(opt.Level == 1 || opt.Key == "Info" || opt.Level == 2 || opt.Key == "Warn") {
	//	callerInfo = CallerInfo()
	//}
	// Log to all receivers
	for _, r := range l.Receivers {
		r.log(opt, callerInfo+s)
	}
}

// Debug logs arguments
func (l *DefaultLogger) Debug(a ...interface{}) {
	l.logAll(optDebug, fmt.Sprint(a...))
}

// Info logs arguments
func (l *DefaultLogger) Info(a ...interface{}) {
	l.logAll(optInfo, fmt.Sprint(a...))
}

// Warn logs arguments
func (l *DefaultLogger) Warn(a ...interface{}) {
	l.logAll(optWarn, fmt.Sprint(a...))
}

// Error logs arguments
func (l *DefaultLogger) Error(a ...interface{}) {
	l.logAll(optError, fmt.Sprint(a...))
}

// Fatal logs arguments
func (l *DefaultLogger) Fatal(a ...interface{}) {
	l.logAll(optFatal, fmt.Sprint(a...))
	os.Exit(1)
}

// Panic logs arguments
func (l *DefaultLogger) Panic(a ...interface{}) {
	s := fmt.Sprint(a...)
	l.logAll(optError, s)
	panic(s)
}

// Debugf logs formated arguments
func (l *DefaultLogger) Debugf(format string, a ...interface{}) {
	l.logAll(optDebug, fmt.Sprintf(format, a...))
}

// Infof logs formated arguments
func (l *DefaultLogger) Infof(format string, a ...interface{}) {
	l.logAll(optInfo, fmt.Sprintf(format, a...))
}

// Warnf logs formated arguments
func (l *DefaultLogger) Warnf(format string, a ...interface{}) {
	l.logAll(optWarn, fmt.Sprintf(format, a...))
}

// Errorf logs formated arguments
func (l *DefaultLogger) Errorf(format string, a ...interface{}) {
	l.logAll(optError, fmt.Sprintf(format, a...))
}

// Fatalf logs formated arguments
func (l *DefaultLogger) Fatalf(format string, a ...interface{}) {
	l.logAll(optFatal, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Panicf logs formated arguments
func (l *DefaultLogger) Panicf(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	l.logAll(optError, s)
	panic(s)
}

func (l *DefaultLogger) EnableCounter(labelNames ...string) *prometheus.CounterVec {
	if len(labelNames) == 0 {
		labelNames = append(labelNames, "level")
	}
	l.counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: l.prefix + "log_total",
		Help: "Total number of log items.",
	}, labelNames)
	l.enableCounter = true
	return l.counter
}

// Open is a short function to open a file with needed options
func Open(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
}

// Receiver holds all receiver options
type Receiver struct {
	// DefaultLogger object from the builtin log package
	Logger *log.Logger
	Level  Level
	Color  bool
	Active bool
	Format string
}

// AddReceiver add a receiver to the gLogger's receivers list
func AddReceiver(receiver *Receiver) {
	gLogger.(*DefaultLogger).Receivers = append(gLogger.(*DefaultLogger).Receivers, receiver)
}

// SetPrefix sets the prefix of the gLogger.
// If a prefix is set and no trailing space is written, write one
func (r *Receiver) SetPrefix(prefix string) {
	if prefix != "" && !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	r.Logger.SetPrefix(prefix)
}

// Logs to the gLogger
func (r *Receiver) log(opt *levelOptions, s string) {
	// Don't do anything if not wanted
	if !r.Active || opt.Level < r.Level {
		return
	}
	// Pre- and suffix
	prefix := ""
	suffix := "\n"
	// Add colors if wanted
	if r.Color {
		prefix += fmt.Sprintf("\x1b[0;%sm", strconv.Itoa(opt.Color))
		suffix = "\x1b[0m" + suffix
	}
	// Print to the gLogger
	r.Logger.Printf(prefix+r.Format+suffix, opt.Key, s)
}

// NewReceiver returns a new Receiver object with a given Writer
// and sets default values
func NewReceiver(w io.Writer, prefix string) *Receiver {
	logger := log.New(w, "", log.LstdFlags)
	r := &Receiver{
		Logger: logger,
	}
	// Default options
	r.Active = true
	r.Level = INFO
	r.Format = "[%s] ▶ %s"
	r.SetPrefix(prefix)
	return r
}
