package id

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/snowflake"
	"net"
	"time"
)

type Option func(g *Generator)

func WithSnowFlake(snowflake *snowflake.Snowflake) Option {
	return func(g *Generator) {
		g.snowflake = snowflake
	}
}

func WithRedisClient(client *redis.Client) Option {
	return func(g *Generator) {
		g.client = client
	}
}

func WithRedisTTL(ttl time.Duration) Option {
	return func(g *Generator) {
		g.ttl = ttl
	}
}

const (
	Stopped int = iota
	Running
)

type Generator struct {
	ctx       context.Context
	snowflake *snowflake.Snowflake
	client    *redis.Client
	ttl       time.Duration
	keyFormat string
	uid       string
	status    int
	err       error
}

func NewGenerator(opts ...Option) (g *Generator, err error) {
	uid, err := mac()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	g = &Generator{
		ctx:       context.Background(),
		ttl:       time.Second * 12,
		keyFormat: "/snowflake/worker/%d",
		uid:       uid,
		status:    Stopped,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.client == nil {
		// 默认redis client localhost:6379
		g.client = redis.NewClient(&redis.Options{})
	}

	if g.snowflake == nil {
		// 默认snowflake
		g.snowflake, err = snowflake.NewSnowflake()
	}

	err = g.registerWorkerId()
	if err != nil {
		return nil, err
	}

	return
}

func (g *Generator) Next() (id int64, err error) {
	if g.status != Running {
		return 0, errors.New("Id generator is not running")
	}

	return g.snowflake.Next(), nil
}

func (g *Generator) registerWorkerId() (err error) {
	ctx := context.Background()

	max := int(g.snowflake.GetMaxWorks())
	for i := 0; i < max; i++ {
		cmd := g.client.SetNX(ctx, fmt.Sprintf(g.keyFormat, i), g.uid, g.ttl)
		if cmd.Err() != nil {
			return errors.WithStack(cmd.Err())
		}

		// 成功获取
		if cmd.Val() {
			g.status = Running
			g.err = nil
			g.snowflake.SetWorkerId(int64(i))
			return
		}
	}
	return errors.New(fmt.Sprintf("Can not request the worker id: count of machines over %d", max))
}

func (g *Generator) keepalive() {
	ticker := time.NewTicker(g.ttl - time.Second*2)
	go func() {
		for range ticker.C {
			g.tick()
			g.logErr()
		}
	}()
}

func (g *Generator) tick() {
	defer func() {
		if r := recover(); r != nil {
			g.status = Stopped
			if err, ok := r.(error); ok {
				g.err = errors.WithStack(err)
			}
		}
	}()

	if g.status != Running {
		err := g.registerWorkerId()
		if err != nil {
			g.err = err
			return
		}
	}

	val, err := g.client.GetEx(g.ctx, fmt.Sprintf(g.keyFormat, g.snowflake.GetWorkerId()), g.ttl).Result()

	if err != nil {
		g.status = Stopped
		g.err = errors.WithStack(err)
	} else if val != g.uid {
		g.status = Stopped
		g.err = errors.New("")
	}
}

// TODO: log error
func (g *Generator) logErr() {
}

func (g *Generator) refresh() (err error) {
	return errors.WithStack(err)
}

func mac() (addr string, err error) {
	addrList, err := net.Interfaces()
	if err != nil {
		return "", errors.WithStack(err)
	}

	for _, ifs := range addrList {
		if len(ifs.HardwareAddr) >= 6 {
			return ifs.HardwareAddr.String(), nil
		}
	}

	return "", errors.New("No net interface")
}

var generator *Generator

func ReplaceGlobal(g *Generator) {
	generator = g
}

func Next() (id int64, err error) {
	if generator == nil {
		generator, err = NewGenerator()
		if err != nil {
			return 0, err
		}
	}
	return generator.Next()
}
