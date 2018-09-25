package log

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

// Logstash represents a connection to a Logstash instance
type Logstash struct {
	conn    net.Conn
	Project string // Project name
	Proto   string // TCP or UDP
	Address string // Tcp address
	Env     string // Environment ：Pro Dev Test
}

// NewLogstash creates a new hook to a Logstash instance, which listens on
// `protocol`://`address`.
func NewLogstash(conf *Logstash) (*Logstash, error) {
	conn, err := net.DialTimeout(conf.Proto, conf.Address, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conf.conn = conn
	return conf, nil
}

func (l *Logstash) format(entry *logrus.Entry) ([]byte, error) {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		fields[k] = v
	}
	fields["@timestamp"] = entry.Time.UTC().Format(time.RFC3339)
	fields["message"] = entry.Message
	fields["level"] = entry.Level.String()
	fields["appName"] = l.Project
	fields["env"] = l.Env
	serialized, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func (l *Logstash) Fire(entry *logrus.Entry) error {
	dataBytes, err := l.format(entry)
	if err != nil {
		return err
	}
	if _, err = l.conn.Write(dataBytes); err != nil {
		return err
	}
	return nil
}

func (l *Logstash) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
