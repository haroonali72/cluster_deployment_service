package utils

import (
	"antelope/models"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
	"strings"
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
		return
	}

	cmd := notifier.Client.Publish(ctx.Data.Company+"_"+channel, string(b))

	beego.Info(*cmd)
	published := false
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

				published = true
				break
			} else {

			}
		}
		if !published {
			listName := "L=" +ctx.Data.Company+"_"+ channel
			cmd = notifier.Client.LPush(listName, string(b))
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
func (notifier *Notifier) Subscribe(channel string, ctx Context) *redis.PubSub {
	pubsub := notifier.Client.Subscribe(ctx.Data.Company + "_" + channel)
	return pubsub
}
func (notifier *Notifier) RecieveNotification(channel string, ctx Context, pubsub *redis.PubSub) bool {
	start := time.Now()
	defer pubsub.Close()
	err1 := notifier.Client.Ping()
	fmt.Println(err1)
	for int(time.Since(start).Minutes()) < 1 {
		message, err := pubsub.ReceiveMessage()
		if err != nil {
			return false
		}
		if strings.Contains(message.Payload, "AgentServer") {
			ctx.SendLogs("Agent Notification: "+message.Payload, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			return true
		} else {
			time.Sleep(30 * time.Second)
		}
	}

	return false
}
