package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
	"log"
	"shadowing/database"
	"shadowing/models"
	"strings"
)

type PasswordController struct {
	beego.Controller
}

func (c *PasswordController) LoginForm() {
	admin := c.GetSession("admin")
	if admin == true {
		c.Redirect("/admin", 302)
		return
	}
	c.TplName = "admin/password.html"
}

func (c *PasswordController) Login() {
	username := strings.TrimSpace(c.GetString("firstname"))
	password := strings.TrimSpace(c.GetString("password"))

	var user models.Admin
	if err := database.DB.Where("firstname = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %s", username)
		c.Data["Error"] = "Foydalanuvchi topilmadi!"
		c.TplName = "admin/password.html"
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.Data["Error"] = "Parol noto'g'ri!"
		c.TplName = "admin/password.html"
		return
	}

	c.SetSession("admin", true)
	c.SetSession("user_id", user.ID)
	c.Redirect("/admin", 302)
}
