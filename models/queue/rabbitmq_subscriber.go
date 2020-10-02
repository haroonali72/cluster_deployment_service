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

	url := "amqp://" + beego.AppConfig.String("rabbitmq_user") + ":" + beego.AppConfig.String("rabbitmq_password") + "@" + beego.AppConfig.String("rabbitmq_url") + "/"
	conn, err := amqp.Dial(url)
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
		true,                     // durable
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
	Region         string       `json:"region" description:"Region of the cluster [optional]"`
	Cloud          models.Cloud `json:"cloud" description:"cloud of the cluster [optional]"`
	ManagedCluster models.Cloud `json:"managed_cluster" description:"cloud of the cluster [optional]"`
	ProfileId      string       `json:"profile_id" description:"profile id of the cluster [optional]"`
}

func ProcessWork(task WorkSchema, ctx utils.Context) {

	url := beego.AppConfig.String("raccoon_url") + models.InfraGetEndpoint
	if strings.Contains(url, "{infraId}") {
		url = strings.Replace(url, "{infraId}", task.InfraId, -1)
	}
	data, err := api_handler.GetAPIStatus(task.Token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	var infra Infrastructure
	err = json.Unmarshal(data.([]byte), &infra.infrastructureData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	if infra.infrastructureData.Cloud == models.OP && infra.infrastructureData.ManagedCluster == "" {

		if task.Action == models.Create {

			go OPClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go OPClusterTerminateHelper(task, infra)

		}
	}else if infra.infrastructureData.Cloud == models.AWS && infra.infrastructureData.ManagedCluster == "" {

		if task.Action == models.Create {

			go AWSClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go AWSClusterTerminateHelper(task, infra)

		}
	} else if infra.infrastructureData.Cloud == models.Azure && infra.infrastructureData.ManagedCluster == "" {

		if task.Action == models.Create {

			go AzureClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go AzureClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.GCP && infra.infrastructureData.ManagedCluster == "" {

		if task.Action == models.Create {

			go GCPClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go GCPClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.DO && infra.infrastructureData.ManagedCluster == "" {

		if task.Action == models.Create {

			go DOClusterStartHelper(task, infra)

		} else if task.Action == models.Terminate {

			go DOClusterTerminateHelper(task, infra)
		}
	} else if infra.infrastructureData.Cloud == models.Azure && infra.infrastructureData.ManagedCluster == models.AKS {

		if task.Action == models.Create {

			go AKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			go AKSClusterTerminateHelper(task, infra)

		} else if task.Action == models.Update {

			UpdateAKSRunningCluster(task, infra)

		}
	} else if infra.infrastructureData.Cloud == models.AWS && infra.infrastructureData.ManagedCluster == models.EKS {

		if task.Action == models.Create {

			go EKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			go EKSClusterTerminateHelper(task, infra)

		} else if task.Action == models.Update {

			UpdateEKSRunningCluster(task, infra)

		}
	} else if infra.infrastructureData.Cloud == models.DO && infra.infrastructureData.ManagedCluster == models.DOKS {

		if task.Action == models.Create {

			go DOKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			go DOKSClusterTerminateHelper(task, infra)

		} else if task.Action == models.Update {

			UpdateDOKSRunningCluster(task, infra)

		}
	} else if infra.infrastructureData.Cloud == models.GCP && infra.infrastructureData.ManagedCluster == models.GKE {

		if task.Action == models.Create {

			GKEClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			GKEClusterTerminateHelper(task, infra)

		} else if task.Action == models.Update {

			UpdateGKERunningCluster(task, infra)

		}
	} else if infra.infrastructureData.Cloud == models.IBM && infra.infrastructureData.ManagedCluster == models.IKS {

		if task.Action == models.Create {

			IKSClusterStartHelpler(task, infra)

		} else if task.Action == models.Terminate {

			IKSClusterTerminateHelper(task, infra)

		} else if task.Action == models.Update {

			UpdateIKSRunningCluster(task, infra)

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
