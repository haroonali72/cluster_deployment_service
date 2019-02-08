package aws

import (
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

/*var (
	redisHost    = beego.AppConfig.String("redis_url")
)*/

type Notifier struct {
	Client    *redis.Client
	redisHost string
}

func (notifier *Notifier) notify(channel, status string) {

	cmd := notifier.Client.Publish(channel, status)
	beego.Info(*cmd)
}

func (notifier *Notifier) init_notifier() error {
	if notifier.Client != nil {
		return nil
	}
	notifier.redisHost = beego.AppConfig.String("redis_url")
	options := redis.Options{}
	options.Addr = notifier.redisHost
	notifier.Client = redis.NewClient(&options)

	return nil
}
