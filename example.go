package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// r.GET("/ping", ping)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})
	router.GET("/user/:name/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		c.String(http.StatusOK, message)
	})
	router.POST("/user/:name/*action", func(c *gin.Context) {
		b := c.FullPath() == "/user/:name/*action"
		c.String(http.StatusOK, "%t", b)
	})
	router.GET("/user/groups", func(c *gin.Context) {
		c.String(http.StatusOK, "The available groups are [...]")
	})
	router.Run()
}

// func ping(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// }
