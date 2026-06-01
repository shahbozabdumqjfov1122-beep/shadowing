package database

import "shadowing/models"
import "golang.org/x/crypto/bcrypt"

func SeedUserAdmin() {
	var user models.Admin
	if err := DB.Where("role = ?", "admin").First(&user).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
		user.Firstname = "admin"
		user.Role = "admin"
		user.Password = string(hash)
		DB.Create(&user)
	}
}
