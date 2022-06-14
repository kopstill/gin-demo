package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "PFoxMW0#z0aGlr6VNa%48wBA&&^MTn7r",
})

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

	// ############# redis test start #############
	router.POST("/redis", func(c *gin.Context) {
		err := rdb.Set(ctx, "hello", "world", time.Minute).Err()
		if err != nil {
			panic(err)
		}
	})
	// ############# redis test end #############

	router.Run()
}

// func ping(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// }
