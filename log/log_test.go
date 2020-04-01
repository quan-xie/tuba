package log

import (
	"testing"
)

func init() {
	cfg := &Config{
		LogPath: "/Users/xiequan/data/logs/",
		AppName: "test",
		Debug:   true,
	}
	Init(cfg)
}

func Test_Debug(t *testing.T) {
	Debug("hello debug")
	Debugf("hello number=%d", 100)
}

func Test_Info(t *testing.T) {
	Info("hello")
	Infof("hello number=%d", 100)
}

func Test_Warn(t *testing.T) {
	Warn("hello")
	Warnf("hello  number=%d", 100)
}

func Test_Error(t *testing.T) {
	Error("hello")
	Errorf("hello number=%d", 100)
}

func Test_Fatal(t *testing.T) {
	Fatal("hello")
	Fatalf("hello  number=%d", 100)
}
