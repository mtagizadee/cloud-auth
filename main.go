package main

import "github.com/gin-gonic/gin"


	

func main() {
	r := gin.Default()
	
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/ping", ping)
	r.Run("localhost:3001") 
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}