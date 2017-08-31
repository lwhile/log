package log

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"path"

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
	err := AddSentryHook(sentryDSN, InfoLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 1")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 1")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 0")
	}

	// test add a rotate hook
	err = AddRotateHook("log.log", time.Second, time.Second, "", ErrorLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 1 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 2")
	}

	// test add a async sentry  hook
	err = AddAsyncSentryHook(sentryDSN, DebugLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 3 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 3")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 3 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 3")
	}

	err = AddAsyncSentryHook(sentryDSN, ErrorLevel)
	if err != nil {
		t.Fatal(err)
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 2")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 4 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 4")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 4 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 4")
	}

	AddAsyncGraylogHook("", 0, nil, InfoLevel)
	if len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 3 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.InfoLevel]) != 3")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 5 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.ErrorLevel]) != 5")
	}

	if len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 5 {
		t.Fatal("len(baseLogger.entry.Logger.Hooks[logrus.FatalLevel]) != 5")
	}

}

// func Test_convert2logrusLevels(t *testing.T) {
// 	levels := []Level{InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
// 	lglels := convert2logrusLevels(levels...)

// 	if reflect.TypeOf(lglels[0]) != reflect.TypeOf(logrus.InfoLevel) {
// 		t.Fatal()
// 	}

// 	if reflect.TypeOf(lglels[1]) != reflect.TypeOf(logrus.WarnLevel) {
// 		t.Fatal()
// 	}

// 	if reflect.TypeOf(lglels[2]) != reflect.TypeOf(logrus.FatalLevel) {
// 		t.Fatal()
// 	}
// }

func Test_createDir(t *testing.T) {
	targetPath := path.Join("/tmp", fmt.Sprintf("%d", time.Now().Unix()), fmt.Sprintf("%d", time.Now().Unix())+".log")
	log.Println("targetPath:", targetPath)
	err := createDir(targetPath)
	if err != nil {
		t.Fatal(err)
	}

	targetDir := path.Dir(targetPath)
	err = os.RemoveAll(targetDir)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_higHerLevel(t *testing.T) {
	ls1 := higHerLevel(DebugLevel)
	if len(ls1) != 6 {
		t.Fatalf("len(ls):%d != 6", len(ls1))
	}

	if ls1[0] != DebugLevel {
		t.Fatalf("%s != %s\n", ls1[0], DebugLevel)
	}

	if ls1[1] != InfoLevel {
		t.Fatalf("%s != %s\n", ls1[1], InfoLevel)
	}

	if ls1[2] != WarnLevel {
		t.Fatalf("%s != %s\n", ls1[2], WarnLevel)
	}

	if ls1[3] != ErrorLevel {
		t.Fatalf("%s != %s\n", ls1[3], ErrorLevel)
	}

	if ls1[4] != FatalLevel {
		t.Fatalf("%s != %s\n", ls1[4], FatalLevel)
	}
	if ls1[5] != PanicLevel {
		t.Fatalf("%s != %s\n", ls1[5], PanicLevel)
	}

	ls2 := higHerLevel(FatalLevel)
	if len(ls2) != 2 {
		t.Fatalf("len(ls2) = %d != %d\n", len(ls2), 2)
	}
	if ls2[0] != FatalLevel {
		t.Fatalf("%s != %s\n", ls2[0], FatalLevel)
	}
	if ls2[1] != PanicLevel {
		t.Fatalf("%s != %s\n", ls2[1], PanicLevel)
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
