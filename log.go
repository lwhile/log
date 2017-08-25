// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bytes"

	"github.com/evalphobia/logrus_sentry"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type levelFlag string

const (
	// ForverDay mean 1000 years
	ForverDay = 1000 * 365

	// ForverHour mean 1000 years
	ForverHour = ForverDay * 24
)

// String implements flag.Value.
func (f levelFlag) String() string {
	return fmt.Sprintf("%q", string(f))
}

// The Formatter interface is used to implement a custom Formatter. It takes an
// `Entry`. It exposes all the fields, including the default ones:
//
// * `entry.Data["msg"]`. The message passed from Info, Warn, Error ..
// * `entry.Data["time"]`. The timestamp.
// * `entry.Data["level"]. The level the entry was logged at.
//
// Any additional fields added with `WithField` or `WithFields` are also in
// `entry.Data`. Format is expected to return an array of bytes which are then
// logged to `logger.Out`.
type Formatter interface {
	logrus.Formatter
}

// PFormatter :
type PFormatter struct {
}

// Format implement Formatter interface
// time [Level][source] message
func (p *PFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	// write prefix message to buffer
	preFixFmt := "%s [%s][%s] "
	prefix := fmt.Sprintf(preFixFmt, entry.Time.Format("2006-01-02 15:04:05"), entry.Level.String(), entry.Data["source"].(string))
	if _, err := b.WriteString(prefix); err != nil {
		return nil, err
	}

	// write log content to buffer
	if _, err := b.WriteString(entry.Message); err != nil {
		return nil, err
	}

	if _, err := b.WriteString("\n"); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// PrefixedFormatter is a default variable that exported to package user
var PrefixedFormatter = &PFormatter{}

var dftFormatter = &logrus.TextFormatter{DisableColors: true}

// Set implements flag.Value.
func (f levelFlag) Set(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	origLogger.Level = l
	return nil
}

// setSyslogFormatter is nil if the target architecture does not support syslog.
var setSyslogFormatter func(string, string) error

// setEventlogFormatter is nil if the target OS does not support Eventlog (i.e., is not Windows).
var setEventlogFormatter func(string, bool) error

func setJSONFormatter() {
	origLogger.Formatter = &logrus.JSONFormatter{}
}

type logFormatFlag url.URL

// String implements flag.Value.
func (f logFormatFlag) String() string {
	u := url.URL(f)
	return fmt.Sprintf("%q", u.String())
}

// Set implements flag.Value.
func (f logFormatFlag) Set(format string) error {
	u, err := url.Parse(format)
	if err != nil {
		return err
	}
	if u.Scheme != "logger" {
		return fmt.Errorf("invalid scheme %s", u.Scheme)
	}
	jsonq := u.Query().Get("json")
	if jsonq == "true" {
		setJSONFormatter()
	}

	switch u.Opaque {
	case "syslog":
		if setSyslogFormatter == nil {
			return fmt.Errorf("system does not support syslog")
		}
		appname := u.Query().Get("appname")
		facility := u.Query().Get("local")
		return setSyslogFormatter(appname, facility)
	case "eventlog":
		if setEventlogFormatter == nil {
			return fmt.Errorf("system does not support eventlog")
		}
		name := u.Query().Get("name")
		debugAsInfo := false
		debugAsInfoRaw := u.Query().Get("debugAsInfo")
		if parsedDebugAsInfo, err := strconv.ParseBool(debugAsInfoRaw); err == nil {
			debugAsInfo = parsedDebugAsInfo
		}
		return setEventlogFormatter(name, debugAsInfo)
	case "stdout":
		origLogger.Out = os.Stdout
	case "stderr":
		origLogger.Out = os.Stderr
	default:
		return fmt.Errorf("unsupported logger %q", u.Opaque)
	}
	return nil
}

func init() {
	AddFlags(flag.CommandLine)
}

// AddFlags adds the flags used by this package to the given FlagSet. That's
// useful if working with a custom FlagSet. The init function of this package
// adds the flags to flag.CommandLine anyway. Thus, it's usually enough to call
// flag.Parse() to make the logging flags take effect.
func AddFlags(fs *flag.FlagSet) {
	fs.Var(
		levelFlag(origLogger.Level.String()),
		"log.level",
		"Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]",
	)
	fs.Var(
		logFormatFlag(url.URL{Scheme: "logger", Opaque: "stderr"}),
		"log.format",
		`Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"`,
	)
}

// Logger is the interface for loggers used in the Prometheus components.
type Logger interface {
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})

	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	With(key string, value interface{}) Logger

	AddRotateHookByDay(path string, maxAge, rotateDay int, levels ...Level) error
	AddRotateHookByHour(path string, maxAge, rotateHour int, levels ...Level) error

	AddSentryHook(dsn string, levels ...Level) error
	AddSentryHookWithTag(dsn string, tags map[string]string, levels ...Level) error

	AddAsyncSentryHook(dsn string, levels ...Level) error
}

type logger struct {
	entry *logrus.Entry
}

// sourced adds a source field to the logger that contains
// the file name and line where the logging happened.
func (l logger) sourced() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}
	return l.entry.WithField("source", fmt.Sprintf("%s:%d", file, line))
}

// Level represent trigger level of log
type Level logrus.Level

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

func (l Level) String() (s string) {
	switch {
	case l == InfoLevel:
		s = "info"
	case l == ErrorLevel:
		s = "error"
	case l == WarnLevel:
		s = "warn"
	case l == DebugLevel:
		s = "debug"
	case l == PanicLevel:
		s = "panic"
	case l == FatalLevel:
		s = "fatal"
	}
	return
}

var origLogger = newOrigLogger()
var baseLogger = logger{entry: logrus.NewEntry(origLogger)}

func newOrigLogger() *logrus.Logger {
	return &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
}

func convert2logrusLevels(levels ...Level) []logrus.Level {
	ls := make([]logrus.Level, len(levels))
	for i, l := range levels {
		switch {
		case l == InfoLevel:
			ls[i] = logrus.InfoLevel
		case l == ErrorLevel:
			ls[i] = logrus.ErrorLevel
		case l == WarnLevel:
			ls[i] = logrus.WarnLevel
		case l == DebugLevel:
			ls[i] = logrus.DebugLevel
		case l == PanicLevel:
			ls[i] = logrus.PanicLevel
		case l == FatalLevel:
			ls[i] = logrus.FatalLevel
		}
	}
	return ls
}

// AddSentryHook will add a sentry hook to baseLogger
func AddSentryHook(dsn string, levels ...Level) error {
	return addSentryHook(baseLogger, dsn, levels...)
}

func addSentryHook(l logger, dsn string, levels ...Level) error {
	ls := convert2logrusLevels(levels...)
	hook, err := logrus_sentry.NewSentryHook(dsn, ls)
	if err != nil {
		return err
	}
	l.entry.Logger.Hooks.Add(hook)
	return nil
}

// AddAsyncSentryHook will add a async sentry hook to base logger
func AddAsyncSentryHook(dsn string, levels ...Level) error {
	return addAsyncSentryHook(baseLogger, dsn, levels...)
}

func addAsyncSentryHook(l logger, dsn string, levels ...Level) error {
	ls := convert2logrusLevels(levels...)
	hook, err := logrus_sentry.NewAsyncSentryHook(dsn, ls)
	if err != nil {
		return err
	}
	l.entry.Logger.Hooks.Add(hook)
	return nil
}

// AddSentryHookWithTag wii add a sentry hook with tag to baseLogger
func AddSentryHookWithTag(dsn string, tags map[string]string, levels ...Level) error {
	return addSentryHookWithTag(baseLogger, dsn, tags, levels...)
}

func addSentryHookWithTag(l logger, dsn string, tags map[string]string, levels ...Level) error {
	ls := convert2logrusLevels(levels...)
	hook, err := logrus_sentry.NewWithTagsSentryHook(dsn, tags, ls)
	if err != nil {
		return err
	}
	l.entry.Logger.Hooks.Add(hook)
	return nil
}

// AddRotateHook will add a rotate hook to baseLogger
func AddRotateHook(path string, maxAge, rotateTime time.Duration, format string, levels ...Level) error {
	return addRotateHook(baseLogger, path, maxAge, rotateTime, format, levels...)
}

func addRotateHook(l logger, path string, maxAge, rotateTime time.Duration, format string, levels ...Level) error {
	if err := createDir(path); err != nil {
		return err
	}
	for _, level := range levels {
		writer, err := rotatelogs.New(
			fmt.Sprintf("%s.%s.%s", path, level.String(), format), rotatelogs.WithLinkName(path),
			rotatelogs.WithMaxAge(maxAge),
			rotatelogs.WithRotationTime(rotateTime),
		)
		if err != nil {
			return err
		}

		writeMap := getWriteMap(level, writer)

		hook := lfshook.NewHook(writeMap)
		l.entry.Logger.Hooks.Add(hook)
	}
	return nil
}

// AddRotateHookByDay will add a rotate hook to baseLogger rotating by day
func AddRotateHookByDay(path string, maxAge, rotateDay int, levels ...Level) error {
	return addRotateHookByDay(baseLogger, path, maxAge, rotateDay, dftFormatter, levels...)
}

// AddRotateHookByDayWithFormatter will add a rotate hook with formatter to baseLogger rotating by day
func AddRotateHookByDayWithFormatter(path string, maxAge, rotateDay int, formatter Formatter, levels ...Level) error {
	return addRotateHookByDay(baseLogger, path, maxAge, rotateDay, formatter, levels...)
}

func addRotateHookByDay(l logger, path string, maxAge, rotateDay int, formatter Formatter, levels ...Level) error {
	if err := createDir(path); err != nil {
		return err
	}
	for _, level := range levels {
		writer, err := rotatelogs.New(
			fmt.Sprintf("%s.%s.%s", path, level.String(), "%Y-%m-%d"), rotatelogs.WithLinkName(path),
			rotatelogs.WithMaxAge(time.Duration(maxAge)*time.Hour*24),
			rotatelogs.WithRotationTime(time.Duration(rotateDay)*time.Hour*24),
		)
		if err != nil {
			return err
		}

		writeMap := getWriteMap(level, writer)

		hook := lfshook.NewHook(writeMap)
		hook.SetFormatter(formatter)

		l.entry.Logger.Hooks.Add(hook)
	}
	return nil
}

// AddRotateHookByHour will add a rotate hook to baseLogger rotating by hour
func AddRotateHookByHour(path string, maxAge, rotateHour int, levels ...Level) error {
	return addRotateHookByHour(baseLogger, path, maxAge, rotateHour, levels...)
}

// AddRotateHookByHour will add a rotate hook to baseLogger rotating by hour
func addRotateHookByHour(l logger, path string, maxAge, rotateHour int, levels ...Level) error {
	if err := createDir(path); err != nil {
		return err
	}
	for _, level := range levels {
		writer, err := rotatelogs.New(
			fmt.Sprintf("%s.%s.%s", path, level.String(), "%Y-%m-%d@%H:00"), rotatelogs.WithLinkName(path),
			rotatelogs.WithMaxAge(time.Duration(maxAge)*time.Hour),
			rotatelogs.WithRotationTime(time.Duration(maxAge)*time.Hour),
		)
		if err != nil {
			return err
		}

		writeMap := getWriteMap(level, writer)
		hook := lfshook.NewHook(writeMap)
		l.entry.Logger.Hooks.Add(hook)
	}
	return nil
}

func getWriteMap(level Level, writer *rotatelogs.RotateLogs) lfshook.WriterMap {
	var writeMap lfshook.WriterMap

	switch {
	case level == ErrorLevel:
		writeMap = lfshook.WriterMap{logrus.ErrorLevel: writer}
	case level == WarnLevel:
		writeMap = lfshook.WriterMap{logrus.WarnLevel: writer}
	case level == DebugLevel:
		writeMap = lfshook.WriterMap{logrus.DebugLevel: writer}
	case level == PanicLevel:
		writeMap = lfshook.WriterMap{logrus.PanicLevel: writer}
	case level == FatalLevel:
		writeMap = lfshook.WriterMap{logrus.FatalLevel: writer}
	default:
		writeMap = lfshook.WriterMap{logrus.InfoLevel: writer}
	}

	return writeMap
}

// Base returns the default Logger logging to
func Base() Logger {
	return baseLogger
}

func (l logger) With(key string, value interface{}) Logger {
	return logger{
		entry: l.entry.WithField(key, value),
	}
	//return logger{entry: l.entry.WithField(key, value)}
}

// Debug logs a message at level Debug on the standard logger.
func (l logger) Debug(args ...interface{}) {
	l.sourced().Debug(args...)
}

// Debug logs a message at level Debug on the standard logger.
func (l logger) Debugln(args ...interface{}) {
	l.sourced().Debugln(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (l logger) Debugf(format string, args ...interface{}) {
	l.sourced().Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func (l logger) Info(args ...interface{}) {
	l.sourced().Info(args...)
}

// Info logs a message at level Info on the standard logger.
func (l logger) Infoln(args ...interface{}) {
	l.sourced().Infoln(args...)
}

// Infof logs a message at level Info on the standard logger.
func (l logger) Infof(format string, args ...interface{}) {
	l.sourced().Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l logger) Warn(args ...interface{}) {
	l.sourced().Warn(args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l logger) Warnln(args ...interface{}) {
	l.sourced().Warnln(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (l logger) Warnf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func (l logger) Error(args ...interface{}) {
	l.sourced().Error(args...)
}

// Error logs a message at level Error on the standard logger.
func (l logger) Errorln(args ...interface{}) {
	l.sourced().Errorln(args...)
}

// Errorf logs a message at level Error on the standard logger.
func (l logger) Errorf(format string, args ...interface{}) {
	l.sourced().Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l logger) Fatal(args ...interface{}) {
	l.sourced().Fatal(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l logger) Fatalln(args ...interface{}) {
	l.sourced().Fatalln(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (l logger) Fatalf(format string, args ...interface{}) {
	l.sourced().Fatalf(format, args...)
}

func (l logger) AddSentryHook(dsn string, levels ...Level) error {
	return addSentryHook(l, dsn, levels...)
}

func (l logger) AddSentryHookWithTag(dsn string, tags map[string]string, levels ...Level) error {
	return addSentryHookWithTag(l, dsn, tags, levels...)
}

func (l logger) AddRotateHookByDay(path string, maxAge, rotateDay int, levels ...Level) error {
	return addRotateHookByDay(l, path, maxAge, rotateDay, dftFormatter, levels...)
}

func (l logger) AddRotateHookByHour(path string, maxAge, rotateHour int, levels ...Level) error {
	return addRotateHookByHour(baseLogger, path, maxAge, rotateHour, levels...)
}

func (l logger) AddAsyncSentryHook(dsn string, levels ...Level) error {
	return addAsyncSentryHook(l, dsn, levels...)
}

// NewLogger returns a new Logger logging to out.
func NewLogger(w io.Writer) Logger {
	l := logrus.New()
	l.Out = w
	return logger{entry: logrus.NewEntry(l)}
}

// NewNopLogger returns a logger that discards all log messages.
func NewNopLogger() Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return logger{entry: logrus.NewEntry(l)}
}

// With adds a field to the logger.
func With(key string, value interface{}) Logger {
	return baseLogger.With(key, value)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	baseLogger.sourced().Debug(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	baseLogger.sourced().Debugln(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	baseLogger.sourced().Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	baseLogger.sourced().Info(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	baseLogger.sourced().Infoln(args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	baseLogger.sourced().Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	baseLogger.sourced().Warn(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	baseLogger.sourced().Warnln(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	baseLogger.sourced().Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	baseLogger.sourced().Error(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	baseLogger.sourced().Errorln(args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	baseLogger.sourced().Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	baseLogger.sourced().Fatal(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	baseLogger.sourced().Fatalln(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	baseLogger.sourced().Fatalf(format, args...)
}

type errorLogWriter struct{}

func (errorLogWriter) Write(b []byte) (int, error) {
	baseLogger.sourced().Error(string(b))
	return len(b), nil
}

// NewErrorLogger returns a log.Logger that is meant to be used
// in the ErrorLog field of an http.Server to log HTTP server errors.
func NewErrorLogger() *log.Logger {
	return log.New(&errorLogWriter{}, "", 0)
}

func createDir(filePath string) error {
	dir := path.Dir(filePath)
	if dir == "" {
		return fmt.Errorf("Failed to create dir")
	}

	return os.MkdirAll(dir, 0770)
}
