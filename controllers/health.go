package controllers

import (
	"github.com/astaxie/beego"
)

// health endpoint
type HealthController struct {
	beego.Controller
}

// @Title Get
// @Description get health
// @Success 200 "alive!"
// @router / [get]
func (c *HealthController) Get() {
	c.Data["json"] = "alive!"
	c.ServeJSON()
}