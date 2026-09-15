package main

import (
	"fmt"
	"os"
	"strconv"

	//"net/http"
	conf "github.com/Tejeda009/goapi/config"
	db "github.com/Tejeda009/goapi/database"
	auth "github.com/Tejeda009/goapi/middleware"
	sec "github.com/Tejeda009/goapi/middleware/security"
	route "github.com/Tejeda009/goapi/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading the env file")
	}

	conf.Key = []byte(os.Getenv("JWT_KEY"))
	conf.Iss = os.Getenv("JWT_ISS")

	costr := os.Getenv("BCRYPT_COST")
	conf.BcryptCost, err = strconv.Atoi(costr)
	if err != nil || conf.BcryptCost < 4 || conf.BcryptCost > 31 {
		conf.BcryptCost = 10
	}

	db.StartDB()

	db.AddAdmin()

	router := gin.Default()
	// gin.SetMode(gin.ReleaseMode) production only

	router.Use(sec.SecurityHeaders())
	router.Use(sec.RateLimiter())

	router.POST("/login", auth.Login)

	{
		var dec string
		authorized := router.Group("/", sec.Auth(&dec))

		authorized.PUT("/lists/:name", route.UpdateName)

		authorized.POST("/lists/:name", route.CreateShoppingList)
		authorized.POST("/lists/:name/items", route.AppendItems)
		authorized.POST("/lists/:name/items/update", route.UpdateItem)

		authorized.GET("/lists", route.GetShoppingLists)
		authorized.GET("/lists/:name", route.GetItems)

		authorized.DELETE("/lists/:name", route.RemoveShoppingList)
		authorized.DELETE("/lists/:name/items", route.RemoveItems) // remove items of a list

		authorized.POST("admin/register", auth.Register)
	}

	err = router.Run(":8090")
	router.Use(gin.Recovery())
	router.Use(gin.ErrorLogger())

	if err != nil {
		panic("Error starting the API")
	}
}
