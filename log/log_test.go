package log

import (
	"testing"

	"go.uber.org/zap"
)


func init(){
	cfg:=&Config{
		Development:true,
		OutputPaths:[]string{"stdout","/data/logs/tuba.log"},
		ErrorOutputPaths:[]string{"stderr"},
	}
	Init(cfg)
}

func Test_Info(t *testing.T) {
	Info("hello",zap.Any("%d",10))
}

func Test_Error(t *testing.T) {
	Error("hello error",zap.Any("number",10))
}