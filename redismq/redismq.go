package redismq

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/vuuvv/errors"
	"github.com/vuuvv/orca/serialize"
	"github.com/vuuvv/orca/utils"
	"go.uber.org/zap"
	"strings"
	"time"
)

func groupName(queue string) string {
	return fmt.Sprintf("%s_group", queue)
}

func consumerName(queue string) string {
	return fmt.Sprintf("%s_group_consumer", queue)
}

func Produce(cli *redis.Client, queue string, value interface{}) error {
	body, err := serialize.JsonStringify(value)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = cli.XAdd(context.Background(), &redis.XAddArgs{
		Stream: queue,
		MaxLen: 1000,
		Approx: true,
		Values: []string{"payload", string(body)},
	}).Result()
	return errors.WithStack(err)
}

func Consume(cli *redis.Client, queue string, handler func(payload string) error) {
	defer utils.NormalRecover("Consume")

	group := groupName(queue)
	consumer := consumerName(queue)

	for {
		payload, err := cli.XReadGroup(context.Background(), &redis.XReadGroupArgs{
			Streams:  []string{queue, ">"},
			Group:    group,
			Consumer: consumer,
			Count:    1,
			Block:    0,
		}).Result()
		if err != nil {
			if strings.HasPrefix(err.Error(), "NOGROUP") {
				_, err = cli.XGroupCreateMkStream(context.Background(), queue, group, "0").Result()
				if err == nil {
					continue
				}
			}
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Error(err))
			time.Sleep(time.Second * 2)
			continue
		}
		if len(payload) == 0 {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Error(errors.New("消息为空")))
			continue
		}
		messages := payload[0].Messages
		if len(messages) == 0 {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Any("payload", payload), zap.Error(errors.New("消息为空")))
			continue
		}
		msg := messages[0].Values
		if messages[0].Values == nil {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Any("payload", payload), zap.Error(errors.New("消息为空")))
			continue
		}
		body, ok := msg["payload"]
		if !ok {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Any("payload", payload), zap.Error(errors.New("消息为空")))
			continue
		}
		err = handler(body.(string))
		if err != nil {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Any("payload", payload), zap.Error(err))
			continue
		}
		_, err = cli.XAck(context.Background(), queue, group, messages[0].ID).Result()
		if err != nil {
			zap.L().Error("Consume queue error", zap.String("queue", queue), zap.Any("payload", payload), zap.Error(err))
			continue
		}
	}
}
