package utils

import (
	"antelope/models"
	"antelope/models/types"
	"encoding/json"
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
	Data    SDData
}

func (c *Context) SendLogs(message, severity string, logType models.Logger) {
	switch logType {
	case models.Backend_Logging:

		c.Data.LogName = string(models.Backend_Logging)

		_, file, line, _ := runtime.Caller(1)
		c.Data.Message = file + ":" + strconv.Itoa(line) + " " + message

		go c.Log(message, severity, logType)

	case models.Audit_Trails:
		c.Data.LogName = string(models.Audit_Trails)
		c.Data.Message = message + " by User: " + c.Data.UserId
		go c.Log(message, severity, logType)

	}

}

func (c *Context) Log(msg, message_type string, logType models.Logger) (int, error) {

	c.Data.Severity = message_type

	if message_type == models.LOGGING_LEVEL_ERROR {

		c.Data.MessageType = "stderr"
		beego.Error(c.Data.Message)

	} else if message_type == models.LOGGING_LEVEL_INFO {

		c.Data.MessageType = "stdout"
		beego.Info(c.Data.Message)
	}

	logger := InitReq()

	request_data, err := TransformData(c.Data)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	req, err := CreatePostRequest(request_data, c.getHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	SetHeaders(req, m)
	response, err := logger.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	beego.Info(response.StatusCode)
	return response.StatusCode, err

}

func (c *Context) InitializeLogger(requestURL, method, path string, projectId string, companyId string, userId string) {

	c.Data.ResourceName = "Cluster"
	c.Data.ServiceName = "antelope"
	c.Data.Request.Url = requestURL
	c.Data.Request.Method = method
	c.Data.Request.Path = path
	c.Data.Request.RequestId = uuid.New().String()
	c.Data.ProjectId = projectId
	c.Data.Company = companyId
	c.Data.UserId = userId

}

func (c *Context) ReqRespData(payload types.ReqResPayload) string {
	bytes, _ := json.Marshal(payload)
	return string(bytes)
}

func (c *Context) getHost() string {
	switch c.Data.LogName {
	case string(models.Backend_Logging):
		s := getBackendLogHost()
		return s
	case string(models.Audit_Trails):
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
