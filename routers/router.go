package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"shadowing/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/password", &controllers.PasswordController{}, "get:LoginForm;post:Login")
	beego.Router("/admin", &controllers.AdminController{}, "get,post:Admin")

	beego.Router("/admin/videos", &controllers.VideoController{}, "get:AdminVideos")
	beego.Router("/admin/video/add", &controllers.VideoController{}, "post:AddVideo")
	beego.Router("/video/edit/:id", &controllers.VideoController{}, "get:Edit")
	beego.Router("/video/update/:id", &controllers.VideoController{}, "post:Update")
	beego.Router("/watch/:id", &controllers.VideoController{}, "get:Watch")
	beego.Router("/upload-recording/:id", &controllers.VideoController{}, "post:UploadRecording")
	beego.Router("/recording/delete/:id", &controllers.VideoController{}, "post:DeleteRecording")
	// Routerni shunday o'zgartiring (boshidan /admin olib tashlandi):
	beego.Router("/audio/add/:id", &controllers.VideoController{}, "post:AddAudio")
	beego.Router("/audio/update/:id", &controllers.VideoController{}, "post:UpdateSingleAudio")
	beego.Router("/audio/delete/:id", &controllers.VideoController{}, "get:DeleteAudio")
	beego.Router("/admin/audio/update/:id", &controllers.VideoController{}, "post:UpdateSingleAudio")
	beego.Router("/admin/audio/delete/:id", &controllers.VideoController{}, "get:DeleteAudio")
	beego.Router("/admin/audio/add/:id", &controllers.VideoController{}, "post:AddAudio")
	beego.Router("/video/upload-recording/:id", &controllers.VideoController{}, "post:UploadRecording")
	beego.Router("/video/recording/delete/:id", &controllers.VideoController{}, "post:DeleteRecording") // 🔥 POST metod bo'lishi shart!
}
