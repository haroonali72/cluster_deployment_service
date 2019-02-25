package utils

import (
	"encoding/json"
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
type Response struct {
	Status    string `json:"status"`
	ID        string `json:"_id"`
	Component string `json:"component"`
}

func (notifier *Notifier) Notify(channel, status string) {
	msg := Response{
		Status:    status,
		ID:        channel,
		Component: "Cluster",
	}
	b, err := json.Marshal(msg)
	if err != nil {
		beego.Error(err.Error())
		return
	}
	cmd := notifier.Client.Publish(channel, string(b))
	beego.Info(*cmd)
	if cmd != nil {
		beego.Error(cmd.Err().Error())
	}
}

func (notifier *Notifier) Init_notifier() error {
	if notifier.Client != nil {
		return nil
	}
	notifier.redisHost = beego.AppConfig.String("redis_url")
	options := redis.Options{}
	options.Addr = notifier.redisHost
	notifier.Client = redis.NewClient(&options)

	return nil
}
