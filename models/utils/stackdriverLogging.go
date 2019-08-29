package utils

import (
	"antelope/models"
	"github.com/astaxie/beego"
	"github.com/google/uuid"
	"runtime"
	"strconv"
)

type SDData struct {
	Request      HTTPRequest `json:"http_request"`
	Message      interface{} `json:"message"`
	MessageType  string      `json:"message_type"`
	ProjectId    string      `json:"project_id"`
	ServiceName  string      `json:"service_name"`
	Severity     string      `json:"severity"`
	UserId       string      `json:"user_id"`
	ResourceName string      `json:"resource_name"` ///??
	Company      string      `json:"company_id" binding:"required"`
	LogName      string      `json:"log_name"`
	Response     interface{} `json:"response"` ///???
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

/*
type AuditTrailRequest struct {

	LogName constants.Logger `json:"log_name"`
	ProjectId string `json:"project_id"`
	ResourceName string `json:"resource_name"`
	ServiceName string `json:"service_name" binding:"required"`
	Severity string `json:"severity" binding:"required"`
	UserId string `json:"user_id" binding:"required"`
	Company string `json:"company_id" binding:"required"`
	MessageType string `json:"message_type"`
	Response interface{} `json:"response"`
	Message      interface{} `json:"message" binding:"required" `
	Http_Request struct {
		Request_Id string `json:"request_id" binding:"required"`
		Url string `json:"url"`
		Method string `json:"method" `
		Path string `json:"path"`
		Body string `json:"body"`
		Status int `json:"status"`
	} `json:"http_request"  binding:"required"`
}
*/

func (c *Context) SendLogs(message, severity string, logType string) {
	switch models.Logger(logType) {
	case models.Backend_Logging:
		c.SendSDLog(message, severity)
	case models.Audit_Trails:
		c.SendAuditTrails(message, severity)

	}
}

func (c *Context) SendAuditTrails(msg, message_type string) (int, error) {
	c.data.LogName = models.Audit_Trail
	msg = msg + "by User: " + c.data.UserId + " of Company: " + c.data.Company
	StatusCode, err := c.Log(msg, message_type)
	return StatusCode, err
}

func (c *Context) SendSDLog(msg, message_type string) (int, error) {
	c.data.LogName = models.Backend_Log
	StatusCode, err := c.Log(msg, message_type)
	return StatusCode, err
}

func (c *Context) Log(msg, message_type string) (int, error) {
	_, file, line, _ := runtime.Caller(1)
	c.data.Severity = message_type
	c.data.Message = file + ":" + strconv.Itoa(line) + " " + msg
	if message_type == models.LOGGING_LEVEL_ERROR {
		c.data.MessageType = "stderr"
	} else if message_type == models.LOGGING_LEVEL_INFO {
		c.data.MessageType = "stdout"
	}
	if c.data.Severity == models.LOGGING_LEVEL_ERROR {
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

	req, err := CreatePostRequest(request_data, getHost(c))
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

func (c *Context) InitializeLogger(requestURL, method, path string, projectId string, companyId string, userId string) {

	c.data.ServiceName = "antelope"
	c.data.Request.Url = requestURL
	c.data.Request.Method = method
	c.data.Request.Path = path
	c.data.Request.RequestId = uuid.New().String()
	c.data.ProjectId = projectId
	//c.data.LogName = "backend-logging"
	c.data.Company = companyId
	c.data.UserId = userId
}

func getHost(c *Context) string {
	switch c.data.LogName {
	case "backend-logging":
		s := getBackendLogHost()
		return s
	case "audit-trails":
		s := getAuditTrailsHost()
		return s
	}
	return "Host Connection Error"
}
func getBackendLogHost() string {

	return beego.AppConfig.String("logger_url") + models.LoggingEndpoint + models.BackEndLoggingURI
}
func getAuditTrailsHost() string {

	return beego.AppConfig.String("logger_url") + models.LoggingEndpoint + models.AuditTrailLoggingURI
}
