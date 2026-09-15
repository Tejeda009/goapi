package config

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" binding:"required" gorm:"unique; not null"`
	Password string `json:"password" binding:"required" gorm:"not null"`
}

type Item struct {
	ID             int    `json:"id" gorm:"primaryKey;autoIncrement"`
	ShoppingListID uint   `json:"shopping_list_id"`
	Name           string `json:"name" binding:"required"`
	Quantity       string `json:"quantity" `
	Note           string `json:"note"`
	Bought         bool   `json:"bought" gorm:"default:false"`
}

type ShoppingList struct {
	gorm.Model
	Name  string `json:"name" binding:"required" gorm:"unique; not null"`
	Day   string `json:"day"`
	Items []Item `json:"items"`
}

type DeleteItemsRequest struct {
	IDS []int `json:"ids" binding:"required"`
}

type UpdateNameRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateItemRequest struct {
	ID       int    `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Quantity string `json:"quantity"`
	Note     string `json:"note"`
	Bought   bool   `json:"bought"`
}

var ListDB *gorm.DB
var UserDB *gorm.DB

var BcryptCost int
var Key []byte
var Iss string

const BcryptLen int = 72
