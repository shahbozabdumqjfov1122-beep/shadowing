package database

import (
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"shadowing/models"
)

var DB *gorm.DB

func InitDB() {

	runmode := beego.AppConfig.DefaultString("runmode", "dev")

	dbHost, _ := beego.AppConfig.String(runmode + "::db_host")
	dbPort, _ := beego.AppConfig.String(runmode + "::db_port")
	dbUser, _ := beego.AppConfig.String(runmode + "::db_user")
	dbPass, _ := beego.AppConfig.String(runmode + "::db_pass")
	dbName, _ := beego.AppConfig.String(runmode + "::db_name")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tashkent",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf(err.Error())
	}
	DB = db

	migrate()
	SeedUserAdmin()

}
func migrate() {
	err := DB.AutoMigrate(
		&models.Admin{},
		&models.Video{},
		&models.VideoAudio{},
		&models.UserRecording{},
		&models.Room{},
		&models.AudioWord{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
