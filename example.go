package main

import (
	"context"
	"fmt"
	"net/http"

	"log"

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

		c.String(http.StatusOK, "ok")
	})

	// ############# Map as querystring or postform parameters #############
	router.POST("/post_map", func(c *gin.Context) {
		ids := c.QueryMap("ids")
		names := c.PostFormMap("names")

		fmt.Printf("ids: %v; names: %v\n", ids, names)

		c.String(http.StatusOK, "ok")
	})

	// ############# Upload files #############
	router.MaxMultipartMemory = 8 << 20 // 8M (default is 32M)
	// Single file
	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("single-file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form file err: %s", err.Error()))
		} else {
			filename := file.Filename
			log.Println(filename)

			c.SaveUploadedFile(file, "/Users/kopever/Desktop/"+filename)

			c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", filename))
		}
	})
	// Multiple files
	router.POST("/upload_multiple", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form files err: %s", err.Error()))
		} else {
			files := form.File["multiple-files"]
			if len(files) == 0 {
				c.String(http.StatusBadRequest, "no files received")
			} else {
				for _, file := range files {
					filename := file.Filename
					log.Println(filename)

					c.SaveUploadedFile(file, "/Users/kopever/Desktop/"+filename)

					c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
				}
			}
		}
	})

	// ############# Grouping routes #############
	// Simple group: v1
	v1 := router.Group("/v1")
	{
		v1.POST("/login", nil)
		v1.POST("/submit", nil)
		v1.POST("/read", nil)
	}

	// Simple group: v2
	v2 := router.Group("/v2")
	{
		v2.POST("/login", nil)
		v2.POST("/submit", nil)
		v2.POST("/read", nil)
	}

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
