package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _db *gorm.DB

func getDB() *gorm.DB {
	return _db
}

func main() {
	r := gin.Default()
	
	api := r.Group("/api")
	v1 := api.Group("/v1")

	dsn := "root:root@tcp(127.0.0.1:3306)/auth?charset=utf8mb4&parseTime=True&loc=Local"
  db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_db = db
	fmt.Println("Connection Opened to Database")

	// Migrate the models

	v1.GET("/ping", ping)
	r.Run("localhost:3001") 
}


func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

