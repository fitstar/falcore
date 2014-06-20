package log

import (
	"errors"
	"log"
)

// This is a simple Logger implementation that
// uses the go log package for output.  It's not
// really meant for production use since it isn't
// very configurable.  It is a sane default alternative
// that allows us to not have any external dependencies.
// Use timber or log4go as a real alternative.
type StdLibLogger struct{}

func NewStdLibLogger() Logger {
	return new(StdLibLogger)
}

var (
	levelStrings = [...]string{"[FNST]", "[FINE]", "[DEBG]", "[TRAC]", "[INFO]", "[WARN]", "[EROR]", "[CRIT]"}
)

func (fl StdLibLogger) Finest(arg0 interface{}, args ...interface{}) {
	fl.Log(FINEST, arg0, args...)
}

func (fl StdLibLogger) Fine(arg0 interface{}, args ...interface{}) {
	fl.Log(FINE, arg0, args...)
}

func (fl StdLibLogger) Debug(arg0 interface{}, args ...interface{}) {
	fl.Log(DEBUG, arg0, args...)
}

func (fl StdLibLogger) Trace(arg0 interface{}, args ...interface{}) {
	fl.Log(TRACE, arg0, args...)
}

func (fl StdLibLogger) Info(arg0 interface{}, args ...interface{}) {
	fl.Log(INFO, arg0, args...)
}

func (fl StdLibLogger) Warn(arg0 interface{}, args ...interface{}) error {
	return fl.Log(WARNING, arg0, args...)
}

func (fl StdLibLogger) Error(arg0 interface{}, args ...interface{}) error {
	return fl.Log(ERROR, arg0, args...)
}

func (fl StdLibLogger) Critical(arg0 interface{}, args ...interface{}) error {
	return fl.Log(CRITICAL, arg0, args...)
}

func (fl StdLibLogger) Log(lvl level, arg0 interface{}, args ...interface{}) (e error) {
	defer func() {
		if x := recover(); x != nil {
			var ok bool
			if e, ok = x.(error); ok {
				return
			}
			e = errors.New("Um... barf")
		}
	}()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		argsNew := append([]interface{}{levelStrings[lvl]}, args...)
		log.Printf("%s "+first, argsNew...)
	case func() string:
		// Log the closure (no other arguments used)
		argsNew := append([]interface{}{levelStrings[lvl]}, first())
		log.Println(argsNew...)
	default:
		// Build a format string so that it will be similar to Sprint
		argsNew := append([]interface{}{levelStrings[lvl]}, args...)
		log.Println(argsNew...)
	}
	return nil
}
