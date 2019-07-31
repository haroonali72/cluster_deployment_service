package utils

import (
	"github.com/astaxie/beego"
	"github.com/google/uuid"
	"runtime"
	"strconv"
)

type SDData struct {
	Company     string      `json:"company"`
	Request     HTTPRequest `json:"http_request"`
	Message     interface{} `json:"message"`
	MessageType string      `json:"message_type"`
	ProjectId   string      `json:"project_id"`
	ServiceName string      `json:"service_name"`
	Severity    string      `json:"severity"`
	UserId      string      `json:"user_id"`
}

type HTTPRequest struct {
	Body      string `json:"body"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	RequestId string `json:"request_id"`
	Status    int64  `json:"status"`
	Url       string `json:"url"`
}
type Context struct {
	context beego.Controller
	data    SDData
}

func (c *Context) SendSDLog(msg, message_type string) (int, error) {

	_, file, line, _ := runtime.Caller(1)

	c.data.Severity = message_type
	c.data.Message = file + ":" + strconv.Itoa(line) + " " + msg

	if c.data.Severity == "error" {
		beego.Error(c.data.Message)
	} else {
		beego.Info(c.data.Message)
	}
	logger := InitReq()

	request_data, err := TransformData(c.data)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := CreatePostRequest(request_data, getHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := logger.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	return response.StatusCode, err

}
func (c *Context) InitializeLogger(requestURL, method, path string, projectId string) {

	c.data.ServiceName = "antelope"
	c.data.Request.Url = requestURL
	c.data.Request.Method = method
	c.data.Request.Path = path
	c.data.Request.RequestId = uuid.New().String()
	c.data.ProjectId = projectId
}
func getHost() string {
	//return "https://dapis.cloudplex.cf/api/v1/backend/logging"
	return "http://" + beego.AppConfig.String("logger_url") + ":3500/elephant/api/v1/backend/logging"
}
