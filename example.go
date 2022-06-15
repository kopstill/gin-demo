package main

import (
	"context"
	"fmt"
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
	// ############# Quick start #############
	// gin.SetMode(gin.ReleaseMode)
	// gin.DefaultWriter = ioutil.Discard
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// ############# Parameters in path #############
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

	// ############# Querystring parameters #############
	router.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname")

		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})

	// ############# Multipart/Urlencoded Form #############
	router.POST("/form_post", func(c *gin.Context) {
		message := c.PostForm("message")
		nick := c.DefaultPostForm("nick", "anonymous")

		c.JSON(http.StatusOK, gin.H{
			"status":  "posted",
			"message": message,
			"nick":    nick,
		})
	})

	// ############# Another example: query + post form #############
	router.POST("/post", func(c *gin.Context) {
		id := c.Query("id")
		page := c.DefaultQuery("page", "0")
		name := c.PostForm("name")
		message := c.PostForm("message")

		fmt.Printf("id: %s; page: %s; name: %s; message: %s\n", id, page, name, message)
	})

	// ############# Redis test #############
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

	router.Run()
}

type redisKVData struct {
	RKey   string `json:"rKey" binding:"required"`
	RValue string `json:"rValue" binding:"required"`
}
