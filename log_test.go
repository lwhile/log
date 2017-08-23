package log

import (
	"fmt"
	"os"
	"testing"
	"time"

	"reflect"

	"github.com/sirupsen/logrus"
)

var logPath = "log.log"

func removeLogFile(fileName string) error {
	return os.Remove(fileName)
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	if err != nil {
		Error(err)
	}
	return true
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

func Test_AddRotateHook(t *testing.T) {

	err := AddRotateHook("log.log", InfoLevel, time.Second*5, time.Second*5, "%Y-%m-%d@%H:%M")
	if err != nil {
		t.Fatal(err)
	}

	timePrefix := time.Now().Format("2006-01-02@15:04")
	fileName := fmt.Sprintf("%s.%s.%s", logPath, "info", timePrefix)

	done := make(chan struct{})

	for {
		select {
		case <-done:
			if !fileExist(fileName) {
				t.Fatal("log file no exist")
			}
			removeLogFile(fileName)
			return
		default:
			Info(testInfoLog)
		}
		time.Sleep(time.Millisecond * 500)
	}

}

func 

// func Test_AddSentryHook(t *testing.T) {
// 	var sentryDSN = "http://634a78c491f54d2e9666e2ff36e0d747:1bd683754e9f4c38838e50f0b2b28d49@192.168.1.100:9000/2"
// 	err := AddSentryHook(sentryDSN, InfoLevel, ErrorLevel)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	Info(testInfoLog)
// 	Error(testErrorLog)
// }

var testInfoLog = "a test log of info level"
var testErrorLog = "a test log of error level"
