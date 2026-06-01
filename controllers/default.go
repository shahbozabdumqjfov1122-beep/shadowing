package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"shadowing/database"
	"shadowing/models"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {

	var videos []models.Video

	database.DB.
		Order("id DESC").
		Find(&videos)

	c.Data["Videos"] = videos

	c.TplName = "index.html"
}
