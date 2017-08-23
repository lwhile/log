package log

import (
	"testing"

	"reflect"

	"github.com/sirupsen/logrus"
)

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

func Test_AddSentryHook(t *testing.T) {
	var sentryDSN = "http://634a78c491f54d2e9666e2ff36e0d747:1bd683754e9f4c38838e50f0b2b28d49@192.168.1.100:9000/2"
	err := AddSentryHook(sentryDSN, InfoLevel, ErrorLevel)
	if err != nil {
		t.Fatal(err)
	}
	Info("a test log of info level")
	Error("a test log of error level")
}
