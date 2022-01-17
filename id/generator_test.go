package id

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"net"
	"testing"
	"time"
)

func TestInterfaces(t *testing.T) {
	addrs, _ := net.Interfaces()
	for _, ifs := range addrs {
		if len(ifs.HardwareAddr) >= 6 {
			t.Logf("%s,%s", ifs.Name, ifs.HardwareAddr)
		}
	}
}

func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second)

	go func() {
		i := 0
		for typ := range ticker.C {
			if i == 2 {
				ticker.Stop()
			}
			fmt.Println(typ)
			i++
		}
	}()

	time.Sleep(time.Second * 5)
}

func TestNext(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "192.168.137.100:6379",
	})
	g, err := NewGenerator(WithRedisClient(client))
	if err != nil {
		t.Fatalf("error on NewGenerator: %+v", err)
	}
	val, err := g.Next()
	t.Log(val)
}
