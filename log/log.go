package log

import (
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config is log config .
type Config struct {
	LogPath   string // 日志存放路径
	AppName   string // 应用名称
	Debug     bool   // 是否开启Debug模式
	MultiFile bool   // 多文件模式根据日志级别生成文件
}

// Init initialize a log config .
func Init(c *Config) {

	var cores []zapcore.Core
	if c.Debug {
		cores = append(cores, zapcore.NewCore(encoder, consoleDebugging, lowPriority))
	} else {
		cores = append(cores, zapcore.NewCore(encoder, consoleDebugging, highPriority))
	}
	if c.LogPath != "" {
		if c.MultiFile {
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+"_info.log")), infoLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+"_warn.log")), warnLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+"_error.log")), errorLevel))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+"_fatal.log")), fatalLevel))
			if c.Debug {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+"_debug.log")), debugLevel))
			}
		} else {
			if c.Debug {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+".log")), lowPriority))
			} else {
				cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(getWriter(c.LogPath+c.AppName+".log")), highPriority))
			}
		}
	}
	setLogger(zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.Development(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.WarnLevel),
	).Sugar())
}

// TimeEncoder time encoder .
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(strconv.FormatInt(time.Now().Unix(), 10))
}

// Debugf log
func Debugf(msg string, args ...interface{}) {
	logger().Debugf(msg, args...)
}

// Debug log
func Debug(args ...interface{}) {
	logger().Debug(args...)
}

// Infof log
func Infof(msg string, args ...interface{}) {
	logger().Infof(msg, args...)
}

// Info log
func Info(args ...interface{}) {
	logger().Info(args...)
}

// Errorf log
func Errorf(msg string, args ...interface{}) {
	logger().Errorf(msg, args...)
}

// Error log
func Error(args ...interface{}) {
	logger().Error(args...)
}

// Warnf log
func Warnf(msg string, args ...interface{}) {
	logger().Warnf(msg, args...)
}

// Warn log
func Warn(args ...interface{}) {
	logger().Warn(args...)
}

// Fatalf send log fatalf
func Fatalf(msg string, args ...interface{}) {
	logger().Fatalf(msg, args...)
}

// Fatal send log fatal
func Fatal(args ...interface{}) {
	logger().Fatal(args...)
}

func getWriter(filename string) io.Writer {
	hook, err := rotatelogs.New(
		strings.Replace(filename, ".log", "", -1)+"_%Y%m%d%H.log",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*7),  // 默认保存时间为7天
		rotatelogs.WithRotationTime(time.Hour), // 每小时滚动一次存储
	)
	if err != nil {
		panic(err)
	}
	return hook
}

func logger() *zap.SugaredLogger {
	return (*zap.SugaredLogger)(atomic.LoadPointer(&loggerImpl))
}

func setLogger(l *zap.SugaredLogger) {
	atomic.StorePointer(&loggerImpl, unsafe.Pointer(l))
}

var loggerImpl unsafe.Pointer = unsafe.Pointer(
	zap.New(zapcore.NewTee(
		zapcore.NewCore(encoder, consoleDebugging, lowPriority)),
		zap.AddCaller(),
		zap.Development(),
		zap.AddCallerSkip(1),
	).Sugar())
