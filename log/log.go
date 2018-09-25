package log

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var xlog = logrus.New()

type Config struct {
	Dir      string
	Logstash *Logstash
}

// Init is Initial log config .
func Init(conf *Config) {
	if conf.Dir != "" {
		pathMap := PathMap{
			logrus.InfoLevel:  conf.Dir + "/info.log",
			logrus.ErrorLevel: conf.Dir + "/error.log",
		}
		xlog.Hooks.Add(NewLocalHook(
			pathMap,
			&logrus.JSONFormatter{},
		))
	}
	logstash, err := NewLogstash(conf.Logstash)
	if err == nil {
		xlog.Hooks.Add(logstash)
	}
	xlog.Error("NewLogstash error(%v)", err)
}

// Info send log info to logstash
func Info(format string, args ...interface{}) {
	xlog.SetLevel(logrus.InfoLevel)
	xlog.Infof(format, args)
}

// Error send log error to logstash
func Error(format string, args ...interface{}) {
	xlog.SetLevel(logrus.ErrorLevel)
	xlog.Errorf(format, args)
}

// Warn send log warn to logstash
func Warn(format string, args ...interface{}) {
	xlog.SetLevel(logrus.WarnLevel)
	xlog.Warnf(format, args)
}

// Debug send log debug to logstash
func Debug(format string, args ...interface{}) {
	xlog.SetLevel(logrus.DebugLevel)
	xlog.Debugf(format, args)
}

// Fatal send log fatal to logstash
func Fatal(format string, args ...interface{}) {
	xlog.SetLevel(logrus.FatalLevel)
	xlog.Fatalf(format, args)
}

// Logger is the logrus logger handler
func Ginlog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknow"
		}
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := logrus.NewEntry(xlog).WithFields(logrus.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(time.RFC3339), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
