package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	conf "github.com/Tejeda009/goapi/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Login(c *gin.Context) {
	var json conf.User
	var user conf.User

	err := c.ShouldBindJSON(&json)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := conf.UserDB.Where("username = ?", json.Username).First(&user)

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
	if err == nil {

		token := jwt.NewWithClaims(jwt.SigningMethodHS512,
			jwt.MapClaims{
				"iss":  conf.Iss,
				"user": user.Username,
			})
		var final string
		final, err = token.SignedString(conf.Key)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		final = "Bearer " + final
		c.Header("Authorization", final) // returns the JWT
		c.JSON(http.StatusOK, gin.H{"success": "logged in"})
		return
	}

	// default
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	return
}

func HashPassword(pwd string) (string, error) {
	var hash []byte
	var err error
	hash, err = bcrypt.GenerateFromPassword([]byte(pwd), conf.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Register(c *gin.Context) {
	var user conf.User
	var err error
	var pwdLen int

	pwdLen, err = strconv.Atoi(os.Getenv("PWD_LEN"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(user.Password) < pwdLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password too short"})
		return
	}

	user.Password, err = HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = conf.UserDB.Create(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "user created"})
}
