package utils

import (
	"antelope/models"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/streadway/amqp"
)

type ResponseSchema struct {
	Status  bool          `json:"status"`
	Message string        `json:"message"`
	InfraId string        `json:"infra_id"`
	Token   string        `json:"token"`
	Action  models.Action `json:"action"`
}

func Publisher(response ResponseSchema, ctx Context) {

	bytes, err := json.Marshal(response)
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	url := "amqp://" + beego.AppConfig.String("rabbitmq_user") + ":" + beego.AppConfig.String("rabbitmq_password") + "@" + beego.AppConfig.String("rabbitmq_url") + "/"

	conn, err := amqp.Dial(url)
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		string(models.DoneQueue), // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         bytes,
		})
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	ctx.SendLogs(" Message Sent."+response.Message, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
}
