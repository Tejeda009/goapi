package main

import (
	"errors"
	"net/http"

	//"net/http"
	sec "github.com/Tejeda009/goapi/middleware/security"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	//"errors"
	"gorm.io/driver/sqlite"
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

type shoppingList struct {
	gorm.Model
	ID    int    `json:"id" binding:"required" gorm:"primaryKey;autoincrement"`
	Name  string `json:"name" binding:"required" gorm:"unique; not null"`
	Day   string `json:"day"`
	Items []Item `json:"items"`
}

type DeleteItemsRequest struct {
	IDS []int `json:"ids" binding:"required"`
}

func getShoppingLists(c *gin.Context) {
	var names []string
	result := ListDB.Select("name").Find(&names)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNoContent, gin.H{"error": "no shopping lists found"})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	c.JSON(http.StatusOK, names)
}

func createShoppingList(c *gin.Context) {

}

func appendItems(c *gin.Context) {

}

func getItems(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
	}

	var list shoppingList

	result := ListDB.Where("name = ?", name).First(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}

	c.JSON(http.StatusOK, gin.H{"success": list})

}

func removeShoppingList(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
	}

	result := ListDB.Where("name = ?", name).Delete(&shoppingList{})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}

	c.JSON(http.StatusOK, gin.H{"success": "true"})
}

func removeItems(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name"})
	}

	var ids DeleteItemsRequest

	err = c.ShouldBindJSON(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
	}

	for _, id := range ids.IDS {
		result := ListDB.Where("id = ?", id).Delete(&Item{})

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "id not found"})
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": "true"})

}

func login(c *gin.Context) {
	var json User
	var user User

	err := c.ShouldBindJSON(&json)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := UserDB.Where("username = ?", json.Username).First(&user)

	// check for general and UserNotFound error
	if result.Error != nil {

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid username or password"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown"})
		return
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(json.Password))
	if err != nil {

		token := jwt.NewWithClaims(jwt.SigningMethodHS512,
			jwt.MapClaims{
				"iss": "GOAPI", // TODO check the claims
				"sub": user.Username,
			})
		final, err := token.SignedString(key)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "issuing the JWT"})
			return
		}

		c.Header("Bearer", final) // returns the JWT
		c.JSON(http.StatusOK, gin.H{"success": "logged in"})
		return
	}

	// default
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	return

}

var ListDB *gorm.DB
var UserDB *gorm.DB
var err error
var key string = "test" // TODO add env

func main() {
	ListDB, err = gorm.Open(sqlite.Open("list.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to list database")
	}

	UserDB, err = gorm.Open(sqlite.Open("users.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to user database")
	}

	err = ListDB.AutoMigrate(&shoppingList{})
	if err != nil {
		panic("Error migrating the list db")
	}

	err = UserDB.AutoMigrate(&User{})
	if err != nil {
		panic("Error migrating the user db")
	}

	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(sec.SecurityHeaders())

	router.POST("/login", login)

	{
		var dec string
		authorized := router.Group("/", sec.Auth(&dec))

		authorized.POST("/lists/:name", createShoppingList)
		authorized.POST("/lists/:name", appendItems)

		authorized.GET("/lists", getShoppingLists)
		authorized.GET("/lists/:name", getItems)

		authorized.DELETE("/lists/:name", removeShoppingList)
		authorized.DELETE("/items/:name", removeItems) // remove items of a list

	}
}
