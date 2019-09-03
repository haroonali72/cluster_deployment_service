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

func (c *Context) SendLogs(message, severity string, logType models.Logger) {
	switch logType {
	case models.Backend_Logging:

		c.data.LogName = string(models.Backend_Logging)

		_, file, line, _ := runtime.Caller(1)
		c.data.Message = file + ":" + strconv.Itoa(line) + " " + message

		go c.Log(message, severity, logType)

	case models.Audit_Trails:
		c.data.LogName = string(models.Audit_Trails)
		c.data.Message = message + " by User: " + c.data.UserId
		go c.Log(message, severity, logType)

	}

}

func (c *Context) Log(msg, message_type string, logType models.Logger) (int, error) {

	c.data.Severity = message_type

	if message_type == models.LOGGING_LEVEL_ERROR {

		c.data.MessageType = "stderr"
		beego.Error(c.data.Message)

	} else if message_type == models.LOGGING_LEVEL_INFO {

		c.data.MessageType = "stdout"
		beego.Info(c.data.Message)
	}

	logger := InitReq()

	request_data, err := TransformData(c.data)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	req, err := CreatePostRequest(request_data, c.getHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := logger.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	beego.Info(response.StatusCode)
	return response.StatusCode, err

}

func (c *Context) InitializeLogger(requestURL, method, path string, projectId string, companyId string, userId string) {

	c.data.ResourceName = "Cluster"
	c.data.ServiceName = "antelope"
	c.data.Request.Url = requestURL
	c.data.Request.Method = method
	c.data.Request.Path = path
	c.data.Request.RequestId = uuid.New().String()
	c.data.ProjectId = projectId
	c.data.Company = companyId
	c.data.UserId = userId

}

func (c *Context) getHost() string {
	switch c.data.LogName {
	case string(models.Backend_Logging):
		return beego.AppConfig.String("logger_url") + models.LoggingEndpoint + models.BackEndLoggingURI
	case string(models.Audit_Trails):
		return beego.AppConfig.String("logger_url") + models.LoggingEndpoint + models.AuditTrailLoggingURI
	}
	return "Host Connection Error"
}
