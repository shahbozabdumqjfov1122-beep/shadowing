package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"shadowing/database"
	"shadowing/models"
)

type AdminController struct {
	beego.Controller
}

func (c *AdminController) Admin() {
	isAdmin := c.GetSession("admin")
	if isAdmin == nil {
		c.Redirect("/password", 302)
		return
	}

	var videos []models.Video

	// 🔥 ENGL MUHIM O'ZGARISH: Videolarni yuklayotganda ularga tegishli ovozlarni ham bazadan preload qilamiz
	database.DB.Preload("Audios").Order("id DESC").Find(&videos)

	c.Data["Videos"] = videos
	c.TplName = "admin/admin.html"
}
