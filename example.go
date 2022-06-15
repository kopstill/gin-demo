package main

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "PFoxMW0#z0aGlr6VNa%48wBA&&^MTn7r",
})

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
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
		var redisKVData redisKVData
		if err := c.ShouldBindJSON(&redisKVData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"bind error": err.Error()})
			return
		}
		if err := rdb.Set(ctx, redisKVData.RKey, redisKVData.RValue, 0).Err(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"redis error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": "0"})
	})
	// ############# redis test end #############

	router.Run()
}

type redisKVData struct {
	RKey   string `json:"rKey" binding:"required"`
	RValue string `json:"rValue" binding:"required"`
}

// func ping(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// }
