package log

import (
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestAddHooks(t *testing.T) {
	// 测试思路:
	// 这里只对 hook 是否成功添加进logrus做测试

	allLevels := []logrus.Level{logrus.ErrorLevel, logrus.InfoLevel, logrus.PanicLevel, logrus.DebugLevel, logrus.WarnLevel, logrus.FatalLevel}
	for _, ele := range allLevels {
		if len(baseLogger.entry.Logger.Hooks[ele]) != 0 {
			t.Fatalf("len(baseLogger.entry.Logger.Hooks[%s]) != 0\n", ele.String())
		}
	}

	// test add a sentry hook
	var sentryDSN = "http://634a78c491f54d2e9666e2ff36e0d747:1bd683754e9f4c38838e50f0b2b28d49@192.168.1.100:9000/2"
	err := AddSentryHook(sentryDSN, InfoLevel, ErrorLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 1")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 1")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 0 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 0")
	}

	// test add a rotate hook
	err = AddRotateHook("log.log", time.Second, time.Second, "", InfoLevel, ErrorLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 0 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 0")
	}

}

func Test_logger_AddHooks(t *testing.T) {

}

func Test_convert2logrusLevels(t *testing.T) {
	levels := []Level{InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	lglels := convert2logrusLevels(levels...)

	if reflect.TypeOf(lglels[0]) != reflect.TypeOf(logrus.InfoLevel) {
		t.Fatal()
	}

	if reflect.TypeOf(lglels[1]) != reflect.TypeOf(logrus.WarnLevel) {
		t.Fatal()
	}

	if reflect.TypeOf(lglels[2]) != reflect.TypeOf(logrus.FatalLevel) {
		t.Fatal()
	}
}

// func Test_AddRotateHook(t *testing.T) {

// }

// func Test_AddSentryHook(t *testing.T) {
// 	var sentryDSN = "http://634a78c491f54d2e9666e2ff36e0d747:1bd683754e9f4c38838e50f0b2b28d49@192.168.1.100:9000/2"
// 	err := AddSentryHook(sentryDSN, InfoLevel, ErrorLevel)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	Info(testInfoLog)
// 	Error(testErrorLog)

// 	hooks := baseLogger.entry.Logger.Hooks
// 	fmt.Println(hooks[logrus.InfoLevel][0])
// }

// var testInfoLog = "a test log of info level"
// var testErrorLog = "a test log of error level"
