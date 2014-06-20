package falcore

import (
	"github.com/fitstar/falcore/log"
	"time"
)

// The "static" package logger
var logger log.Logger = log.NewStdLibLogger()

// Set the packages static logger.  This logger is used by
// falcore can its subpackages.  It is also available through
// the standalone functions that match the logger interface.
// The default is a StdLibLogger
func SetLogger(newLogger log.Logger) {
	logger = newLogger
}

// Helper for calculating times.  return value in Seconds
// DEPRECATED: Use endTime.Sub(startTime).Seconds()
func TimeDiff(startTime time.Time, endTime time.Time) float32 {
	return float32(endTime.Sub(startTime).Seconds())
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Finest(arg0 interface{}, args ...interface{}) {
	logger.Finest(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Fine(arg0 interface{}, args ...interface{}) {
	logger.Fine(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Debug(arg0 interface{}, args ...interface{}) {
	logger.Debug(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Trace(arg0 interface{}, args ...interface{}) {
	logger.Trace(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Info(arg0 interface{}, args ...interface{}) {
	logger.Info(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Warn(arg0 interface{}, args ...interface{}) error {
	return logger.Warn(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Error(arg0 interface{}, args ...interface{}) error {
	return logger.Error(arg0, args...)
}

// Log using the packages default logger.  You can change the
// underlying logger using SetLogger
func Critical(arg0 interface{}, args ...interface{}) error {
	return logger.Critical(arg0, args...)
}
