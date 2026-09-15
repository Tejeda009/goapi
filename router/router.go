package router

import (
	"errors"
	"net/http"

	conf "github.com/Tejeda009/goapi/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetShoppingLists(c *gin.Context) {
	var names []string
	result := conf.ListDB.Select("name").Find(&names)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNoContent, gin.H{"error": "no shopping lists found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, names)
	return
}

func CreateShoppingList(c *gin.Context) {
	name := GetParam("name", c)

	var list conf.ShoppingList

	list.Name = name
	list.Day = ""

	result := conf.ListDB.Save(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name already used"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "list created"})
	return
}

func AppendItems(c *gin.Context) {
	name := GetParam("name", c)

	var items []conf.Item

	err := c.ShouldBindJSON(&items)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var list conf.ShoppingList

	result := conf.ListDB.Where("name = ?", name).First(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	for i := range items {
		items[i].ShoppingListID = list.ID
	}

	result = conf.ListDB.Create(&items)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "items added"})
	return

}

func GetItems(c *gin.Context) {
	name := GetParam("name", c)

	var list conf.ShoppingList

	result := conf.ListDB.Preload("Items").Where("name = ?", name).First(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": list})
	return

}

func RemoveShoppingList(c *gin.Context) {
	name := GetParam("name", c)

	result := conf.ListDB.Where("name = ?", name).Delete(&conf.ShoppingList{})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "true"})
	return
}

func RemoveItems(c *gin.Context) {

	var ids conf.DeleteItemsRequest

	err := c.ShouldBindJSON(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	for _, id := range ids.IDS {
		result := conf.ListDB.Where("id = ?", id).Delete(&conf.Item{})

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "id not found"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": "true"})
	return
}
func GetParam(name string, c *gin.Context) string {
	ret := c.Param(name)

	if ret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid param"})
		return ""
	}

	return ret
}

func UpdateItem(c *gin.Context) {

	var list conf.UpdateItemRequest

	err := c.ShouldBindJSON(&list)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	result := conf.ListDB.Where("id = ?", list.ID).Save(&list)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "item updated"})
	return
}

func UpdateName(c *gin.Context) {
	name := GetParam("name", c)

	var newName conf.UpdateNameRequest

	err := c.ShouldBindJSON(&newName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	var list conf.ShoppingList

	result := conf.ListDB.Where("name = ?", name).First(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "name not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	list.Name = newName.Name
	result = conf.ListDB.Where("name = ?", name).Save(&list)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "name not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "list renamed"})
}
