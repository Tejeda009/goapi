package database

import (
	"fmt"
	"log"
	"os"

	conf "github.com/Tejeda009/goapi/config"
	auth "github.com/Tejeda009/goapi/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddAdmin() {
	var admin conf.User

	admin.Username = os.Getenv("ADMIN_USERNAME")
	if admin.Username == "" {
		fmt.Println("admin user not created, username empty")
		return
	}

	admin.Password = os.Getenv("ADMIN_PASSWORD")
	if admin.Password == "" {
		fmt.Println("admin user not created, password empty")
		return
	}

	if len(admin.Password) > conf.BcryptLen {
		panic("ADMIN password too long")
	}

	var err error
	admin.Password, err = auth.HashPassword(admin.Password)
	if err != nil {
		fmt.Println("admin user not created, error hashing the pwd")
		return
	}

	err = conf.UserDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"password", "updated_at"}),
	}).Create(&admin).Error

	if err != nil {
		log.Println("Internal server error", err.Error())
		return
	}

	fmt.Println("admin user created")
	return
}

func StartDB() {
	var err error
	conf.ListDB, err = gorm.Open(sqlite.Open("list.db?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to list database")
	}

	conf.UserDB, err = gorm.Open(sqlite.Open("users.db?_journal_mode=WAL"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to user database")
	}

	err = conf.ListDB.AutoMigrate(&conf.ShoppingList{}, &conf.Item{})
	if err != nil {
		panic("Error migrating the list db")
	}

	err = conf.UserDB.AutoMigrate(&conf.User{})
	if err != nil {
		panic("Error migrating the user db")
	}
}
