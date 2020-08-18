package utils

import (
	"antelope/models"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/streadway/amqp"
	"log"
)

func Publisher(url string, msg string) {
	conn, err := amqp.Dial(beego.AppConfig.String("rabbitmq_url"))
	if err != nil {
		fmt.Println(err.Error(), "Failed to open a channel")
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err.Error(), "Failed to open a channel")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		string(models.DoneQueue), // name
		false,                    // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		fmt.Println(err.Error(), "Failed to open a channel")
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(msg),
		})
	if err != nil {
		fmt.Println(err.Error(), "Failed to publish a message")
	}
	log.Printf(" [x] Sent %s")
}
