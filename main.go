package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"auth/packages/_jwt"

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
		
	r.POST("/signup", signup)
	r.POST("/login", login)
	r.POST("/verify", verifyToken)

	r.GET("/ping", ping)
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
	AppId int `binding:"required"`
	CompanyId int `binding:"required"`
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
		AppId: dto.AppId,
		CompanyId: dto.CompanyId,
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

type LoginUserDto struct {
	Email string `binding:"required,email"`
	Password string `binding:"required,min=8,max=32"`
}

func (dto *LoginUserDto) ValidatePassword(dbHash string) bool {
	hash := sha256.Sum256([]byte(dto.Password))
	return fmt.Sprintf("%x", hash) == dbHash
}

func login(c *gin.Context) {
	var dto LoginUserDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// find user by email
	var user User
	db := getDB()
	if err := db.Where("email = ?", dto.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// validate password
	if !dto.ValidatePassword(user.Password) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	token, err := _jwt.Token(map[string]int{
		"id": int(user.ID),
	}, 24 * time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "token generation failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
	})
}

func verifyToken(c *gin.Context) {
	h := c.GetHeader("Authorization")
	// split header into 2
	partitions := strings.Split(h, " ")
	if len(partitions) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	// check if the first partition is Bearer
	if partitions[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	// validate token
	claims, err := _jwt.Claims(partitions[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	userId := claims.CustomClaims["id"]
	db := getDB()

	var user User
	if err := db.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	c.JSON(http.StatusOK, user)	
}