package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/quan-xie/tuba/util/xtime"
	xmongo "go.mongodb.org/mongo-driver/mongo"
)

var client *xmongo.Client

func init() {
	cfg := &Config{
		Addrs:         []string{"localhost:55003"},
		Username:      "",
		Password:      "",
		MaxPool:       2,
		ReplicaSet:    "",
		MaxIdletime:   xtime.Duration(time.Second * 10),
		SocketTimeout: xtime.Duration(500 * time.Millisecond),
		ConnTimeout:   xtime.Duration(500 * time.Millisecond),
	}
	client = NewMongo(cfg)
}

func TestDatabase(t *testing.T) {
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Connected to MongoDB!")
}
