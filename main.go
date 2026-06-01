package main

import (
	beego "github.com/beego/beego/v2/server/web"
	"shadowing/database"
	_ "shadowing/routers"
)

func main() {

	beego.SetStaticPath("/uploads", "uploads")

	database.InitDB()

	// 🔥 SHABLON UCHUN "add" FUNKSIYASINI SHU YERDA RO'YXATDAN O'TKAZAMIZ
	beego.AddFuncMap("add", func(x, y int) int {
		return x + y
	})

	beego.Run()
}
