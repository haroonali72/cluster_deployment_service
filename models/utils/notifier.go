package utils

import (
	"antelope/models"
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

func (notifier *Notifier) Notify(channel, status string, ctx Context) {
	msg := Response{
		Status:    status,
		ID:        ctx.data.Company + "_" + channel,
		Component: "Cluster",
	}
	b, err := json.Marshal(msg)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return
	}
	cmd := notifier.Client.Publish(ctx.data.Company+"_"+channel, string(b))
	beego.Info(*cmd)
	//b, err = json.Marshal(*cmd)
	//if err != nil {
	//	beego.Error(err.Error())
	//	return
	//}
	ctx.SendLogs(cmd.String(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	if cmd != nil {
		if cmd.Err() != nil {
			beego.Error(cmd.Err().Error())
		}
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
