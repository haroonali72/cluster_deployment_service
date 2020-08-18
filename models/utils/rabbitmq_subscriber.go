package utils

import (
	"antelope/models"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/streadway/amqp"
	"log"
)

type WorkSchema struct {
	InfraId string `json:"infra_id"`
	token   string `json:"token"`
	Action  string `json:"action"`
}

func Subscriber(url string) {
	conn, err := amqp.Dial(beego.AppConfig.String("rabbitmq_url"))
	if err != nil {
		fmt.Println(err.Error())
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err.Error(), "Failed to open a channel")
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		string(models.WorkQueue), // name
		false,                    // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	forever := make(chan bool)
	var msg []byte
	var task WorkSchema
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			msg = d.Body
			err = json.Unmarshal(d.Body, &task)
			if err != nil {
				fmt.Println(err.Error())
			}
			ProcessWork(task)
			log.Printf("Done")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
func ProcessWork(task WorkSchema) {
	// extra information from middle ware api
	//
	//
}
