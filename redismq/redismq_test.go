package redismq

import (
	"fmt"
	"github.com/vuuvv/orca"
	"sync"
	"testing"
)

func TestProduce(t *testing.T) {
	_ = orca.NewApplication(orca.WithConfigPath("../examples/simple/resources/application.yaml"))
	//ret, err := orca.Redis().Do(context.Background(),  "xadd", "stream_test", "*", "msg", "1").Result()
	for i := 0; i < 100; i++ {
		err := Produce(orca.Redis(), "/test/redismq", map[string]int{"num": i})
		if err != nil {
			t.Log(err)
		}
	}
}

func doConsume(n int, t *testing.T) {
	Consume(orca.Redis(), "/test/redismq", func(payload string) error {
		t.Log(fmt.Sprintf("consumer %d: %s", n, payload))
		return nil
	})
}

var wg sync.WaitGroup

func TestConsume(t *testing.T) {
	_ = orca.NewApplication(orca.WithConfigPath("../examples/simple/resources/application.yaml"))

	wg.Add(1)
	for i := 0; i < 10; i++ {
		go doConsume(i, t)
	}
	wg.Wait()
}
