package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"

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
	err = db.AutoMigrate(&User{})
	if err != nil {
		panic("failed to migrate database")
	}
	fmt.Println("Database Migrated")

	auth := v1.Group("/auth")
	auth.POST("/signup", signup)

	v1.GET("/ping", ping)
	r.Run("localhost:3001") 
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

type User struct {
	gorm.Model
	Email string `gorm:"unique,not null" binding:"required,email"`
	Password string `gorm:"not null" binding:"required,min=8,max=32" json:"-"`
	CompanyId int `gorm:"not null"`
	AppId int `gorm:"not null"`
}

type SignupUserDto struct {
	Email string `binding:"required,email"`
	Password string `binding:"required,min=8,max=32"`
}

func (dto *SignupUserDto) HashPassword() {
	hash := sha256.Sum256([]byte(dto.Password))
	dto.Password = fmt.Sprintf("%x", hash)
}

func signup(c *gin.Context) {
	var dto SignupUserDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	dto.HashPassword()

	user := User{
		Email: dto.Email,
		Password: dto.Password,
	}

	db := getDB()
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, user)	
}
