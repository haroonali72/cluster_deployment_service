package queue

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/streadway/amqp"
	"log"

	"strings"
)

type WorkSchema struct {
	InfraId string        `json:"infra_id"`
	Token   string        `json:"token"`
	Action  models.Action `json:"action"`
	Cloud   models.Cloud  `json:"cloud"`
}

func Subscriber() {

	ctx := new(utils.Context)
	ctx.InitializeLogger("", "GET", "", "", "", "")

	conn, err := amqp.Dial(beego.AppConfig.String("rabbitmq_url"))
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		ctx.SendLogs("Failed to open a channel "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
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
		ctx.SendLogs("Failed to declare a queue "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			ProcessWork(task, *ctx)
			log.Printf("Done")
		}
	}()

	ctx.SendLogs(" [*] Waiting for messages. To exit press CTRL+C", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	<-forever
}

type Infrastructure struct {
	infrastructureData Data_ `json:"data" description:"infrastructure data of the cluster [optional]"`
}
type Data_ struct {
	Region    string       `json:"region" description:"Region of the cluster [optional]"`
	Cloud     models.Cloud `json:"cloud" description:"cloud of the cluster [optional]"`
	ProfileId string       `json:"profile_id" description:"profile id of the cluster [optional]"`
}

func ProcessWork(task WorkSchema, ctx utils.Context) {

	url := beego.AppConfig.String("raccoon_url") + models.InfraGetEndpoint
	if strings.Contains(url, "{InfraId}") {
		url = strings.Replace(url, "{InfraId}", task.InfraId, -1)
	}
	data, err := api_handler.GetAPIStatus(task.Token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}
	var infra Infrastructure
	err = json.Unmarshal(data.([]byte), &infra.infrastructureData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	if infra.infrastructureData.Cloud == models.AWS {

		if task.Action == models.Create {

			go AWSClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go AWSClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.Azure {

		if task.Action == models.Create {

			go AzureClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go AzureClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.GCP {

		if task.Action == models.Create {

			go GCPClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go GCPClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.DO {

		if task.Action == models.Create {

			go DOClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go DOClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.AKS {

		if task.Action == models.Create {

			go AKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			go AKSClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.EKS {

		if task.Action == models.Create {

			go EKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {
			go EKSClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.DOKS {

		if task.Action == models.Create {

		} else if task.Action == models.Terminate {

		}
	} else if infra.infrastructureData.Cloud == models.GKE {

		if task.Action == models.Create {

		} else if task.Action == models.Terminate {

		}
	} else if infra.infrastructureData.Cloud == models.AKS {

		if task.Action == models.Create {

		} else if task.Action == models.Terminate {

		}
	}
}

/*func AgentListener() {

	ctx := new(utils.Context)
	ctx.InitializeLogger("", "GET", "", "", "", "")

	conn, err := amqp.Dial(beego.AppConfig.String("rabbitmq_url"))
	if err != nil {
		ctx.SendLogs("Error in fetching rabbitmq url "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		ctx.SendLogs("Failed to open a channel "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
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
		ctx.SendLogs("Failed to declare a queue "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			ProcessWork(task,*ctx)
			log.Printf("Done")
		}
	}()

	ctx.SendLogs(" [*] Waiting for messages. To exit press CTRL+C", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	<-forever
}*/
