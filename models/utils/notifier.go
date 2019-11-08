package utils

import (
	"antelope/models"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
	"time"
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
		ID:        ctx.Data.Company + "_" + channel,
		Component: "Cluster",
	}

	b, err := json.Marshal(msg)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return
	}

	cmd := notifier.Client.Publish(ctx.Data.Company+"_"+channel, string(b))

	beego.Info(*cmd)

	if cmd != nil && (cmd.Err() != nil || cmd.Val() == 0) {

		start := time.Now()
		for int(time.Since(start).Minutes()) < 1 {

			time.Sleep(5 * time.Second)
			cmd = notifier.Client.Publish(ctx.Data.Company+"_"+channel, string(b))

			if cmd != nil && cmd.Err() != nil {
				beego.Error(cmd.Err().Error())
			} else if cmd != nil && cmd.Val() == 0 {
				beego.Info(*cmd)
			} else if cmd != nil && cmd.Val() > 0 {
				break
			}
		}

		ctx.SendLogs(cmd.String(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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
