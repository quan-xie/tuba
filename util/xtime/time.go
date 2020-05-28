package xtime

import (
	"context"
	"database/sql/driver"
	"strconv"
	"time"
)

type Time int64

func (t *Time) Scan(i interface{}) (err error) {
	switch it := i.(type) {
	case time.Time:
		*t = Time(it.Unix())
	case string:
		var v int64
		v, err = strconv.ParseInt(it, 10, 64)
		*t = Time(v)
	}
	return
}

// Value get time value.
func (t Time) Value() (driver.Value, error) {
	return time.Unix(int64(t), 0), nil
}

// Time get time.
func (t Time) Time() time.Time {
	return time.Unix(int64(t), 0)
}

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration time.Duration

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < time.Duration(d) {
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}
