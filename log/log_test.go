package log

import (
	"testing"
)

func init() {
	cfg := &Config{
		LogPath: "/",
		AppName: "test",
	}
	Init(cfg)
}

func Test_Debug(t *testing.T) {
	Debug("hello %v", 10)
}

func Test_Info(t *testing.T) {
	Info("hello %v", 10)
}

func Test_Warn(t *testing.T) {
	Warn("hello %v", 10)
}

func Test_Error(t *testing.T) {
	Error("hello %v", 10)
}

func Test_Fatal(t *testing.T) {
	Fatal("hello %v", 10)
}
