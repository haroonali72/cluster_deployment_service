package op

import (
	"antelope/models"
	"antelope/models/utils"
	agent_api "bitbucket.org/cloudplex-devs/woodpecker/agent-api"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"strings"
	"sync"
	"time"
)

type AgentConnection struct {
	connection  *grpc.ClientConn
	agentCtx    context.Context
	agentClient agent_api.AgentServerClient
	projectId   string
	companyId   string
	Mux         sync.Mutex
}

func RetryAgentConn(agent *AgentConnection, context2 utils.Context) error {
	err := agent.connection.Close()
	if err != nil {
		context2.SendLogs("error while closing connection :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}
	count := 0
	flag := true
	for flag && count < 5 {
		conn, err := GetGrpcAgentConnection(context2)
		if err != nil {
			count++
		} else {
			agent.connection = conn.connection
			agent.InitializeAgentClient(agent.projectId, agent.companyId)
			flag = false
		}

		time.Sleep(time.Second * 5)
	}

	if count == 5 {
		context2.SendLogs(errors.New("connection cant be established").Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("connection cant be established")
	}
	return nil
}

func GetGrpcAgentConnection(context2 utils.Context) (*AgentConnection, error) {
	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	conn, err := grpc.Dial(beego.AppConfig.String("woodpecker_url"), grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		context2.SendLogs("error while connecting with agent :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &AgentConnection{}, err
	}

	return &AgentConnection{connection: conn}, nil
}
func (agent *AgentConnection) InitializeAgentClient(projectId, companyId string) error {
	if projectId == "" || companyId == "" {
		return errors.New("projectId or companyId must not be empty")
	}
	md := metadata.Pairs(
		"name", *GetAgentID(&projectId, &companyId),
	)
	agent.projectId = projectId
	agent.companyId = companyId
	ctxWithTimeOut, _ := context.WithTimeout(context.Background(), 100*time.Second)
	agent.agentCtx = metadata.NewOutgoingContext(ctxWithTimeOut, md)
	agent.agentClient = agent_api.NewAgentServerClient(agent.connection)
	return nil

}
func (agent *AgentConnection) ExecCommand(node Node, private_key string, ctx utils.Context) error {
	_, err := agent.agentClient.RemoteSSH(agent.agentCtx, &agent_api.SSHRequestUnary{
		Server: node.PrivateIP,
		User:   node.UserName,
		Key:    private_key,
		Cmd:    "ssh",
	})

	if err != nil && (strings.Contains(err.Error(), "all SubConns are in TransientFailure") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "upstream request timeout") || strings.Contains(err.Error(), "transport is closing")) {
		err = RetryAgentConn(agent, ctx)
		if err != nil {
			return err
		}

		_, err := agent.agentClient.RemoteSSH(agent.agentCtx, &agent_api.SSHRequestUnary{
			Server: node.PrivateIP,
			User:   node.UserName,
			Key:    private_key,
			Cmd:    "ssh",
		})
		if err != nil {
			ctx.SendLogs("error while connecting to node :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			ctx.SendLogs("trying again ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

	} else if err != nil {
		ctx.SendLogs("error while connecting to node :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetAgentID(projectId, companyId *string) *string {
	base := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s+%s", *projectId, *companyId)))
	return &base
}
