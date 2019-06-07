package backoff

import (
	"fmt"
	"github.com/quan-xie/tuba/util/xtime"
	"testing"
	"time"
)

func TestExponentialBackoff_Next(t *testing.T) {
	backoff := NewExponentialBackoff(time.Second, 2*time.Second, 2.0)
	for i := 0; i <= 10; i++ {
		fmt.Printf("Interval %d \n", backoff.Next(i))
		time.Sleep(backoff.Next(i))
	}
}

func TestConstantBackoff_Next(t *testing.T) {
	backoff := NewConstantBackoff(xtime.Duration(time.Second))
	for i := 0; i <= 10; i++ {
		fmt.Printf("Interval %d \n", backoff.Next(i))
		time.Sleep(backoff.Next(i))
	}
}
