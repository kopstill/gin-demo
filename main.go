package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v9"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "127.0.0.1:6379",
	Password: "PFoxMW0#z0aGlr6VNa%48wBA&&^MTn7r",
})

func main() {
	// Quick start
	// gin.SetMode(gin.ReleaseMode)
	// gin.DefaultWriter = ioutil.Discard
	// gin.DisableConsoleColor()
	gin.ForceConsoleColor()

	// How to write log file
	f, _ := os.Create("/Users/kopever/Develop/logs/gin-demo/gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.Default()
	// router := gin.New()

	// Custom Log Format
	router.Use(gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			params.ClientIP,
			params.TimeStamp.Format(time.RFC1123),
			params.Method,
			params.Path,
			params.Request.Proto,
			params.StatusCode,
			params.Latency,
			params.Request.UserAgent(),
			params.ErrorMessage,
		)
	}))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Parameters in path
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

	// Querystring parameters
	router.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname")

		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})

	// Multipart/Urlencoded Form
	router.POST("/form_post", func(c *gin.Context) {
		message := c.PostForm("message")
		nick := c.DefaultPostForm("nick", "anonymous")

		c.JSON(http.StatusOK, gin.H{
			"status":  "posted",
			"message": message,
			"nick":    nick,
		})
	})

	// Another example: query + post form
	router.POST("/post", func(c *gin.Context) {
		id := c.Query("id")
		page := c.DefaultQuery("page", "0")
		name := c.PostForm("name")
		message := c.PostForm("message")

		fmt.Printf("id: %s; page: %s; name: %s; message: %s\n", id, page, name, message)

		c.String(http.StatusOK, "ok")
	})

	// Map as querystring or postform parameters
	router.POST("/post_map", func(c *gin.Context) {
		ids := c.QueryMap("ids")
		names := c.PostFormMap("names")

		fmt.Printf("ids: %v; names: %v\n", ids, names)

		c.String(http.StatusOK, "ok")
	})

	// Upload files
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

	// Grouping routes
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

	// Using middleware
	authorized := router.Group("/")
	// authorized.Use(gin.Logger())
	// authorized.Use(gin.Recovery())
	authorized.Use(AuthRequired())
	{
		authorized.POST("/ping", ping())
		authorized.POST("/submit", nil)
		authorized.POST("/read", nil)

		// nested group
		testing := authorized.Group("testing")
		// visit 0.0.0.0:8080/testing/analytics
		testing.GET("/analytics", nil)
	}

	// Custom Recovery behavior
	router.Use(gin.CustomRecovery(func(c *gin.Context, recoverd interface{}) {
		if err, ok := recoverd.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.GET("/panic", func(ctx *gin.Context) {
		panic("foo")
	})
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ohai")
	})

	// Model binding and validation
	router.POST("/loginJSON", func(c *gin.Context) {
		var json Login
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.User != "manu" || json.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	router.POST("/loginXML", func(c *gin.Context) {
		var xml Login
		if err := c.ShouldBindXML(&xml); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if xml.User != "manu" || xml.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	router.POST("/loginForm", func(c *gin.Context) {
		var form Login
		// This will infer what binder to use depending on the content-type header.
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if form.User != "manu" || form.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	// Custom Validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("bookabledate", bookableDate)
	}
	router.GET("/bookable", getBookable)

	// Only Bind Query String
	router.Any("/testing", startPage)

	// Bind Query String or Post Data
	router.Any("/testing1", startPage1)

	// Bind Uri
	router.GET("/:name/:id", func(c *gin.Context) {
		people := People{}
		if err := c.ShouldBindUri(&people); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": people.Name, "uuid": people.ID})
	})

	// Bind Header
	router.GET("/bind_header", func(c *gin.Context) {
		h := testHeader{}
		if err := c.ShouldBindHeader(&h); err != nil {
			c.JSON(http.StatusOK, err)
		}

		fmt.Printf("%#v\n", h)
		c.JSON(http.StatusOK, gin.H{"Rate": h.Rate, "Domain": h.Domain})
	})

	// Bind HTML checkboxes
	router.LoadHTMLFiles("checkbox.html")
	router.GET("/bind_checkbox", checkboxGetHandler)
	router.POST("/bind_checkbox", checkboxPostHandler)

	// Redis test
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

func getBookable(c *gin.Context) {
	var b Book
	if err := c.ShouldBindWith(&b, binding.Query); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Booking dates are valid!"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

type testHeader struct {
	Rate   string `header:"Rate"`
	Domain string `header:"Domain"`
}

type People struct {
	ID   string `uri:"id" binding:"required,uuid"`
	Name string `uri:"name" binding:"required"`
}

type Person struct {
	Name       string    `form:"name"`
	Address    string    `form:"address"`
	Birthday   time.Time `form:"birthday" time_format:"2006-01-02" time_utc:"1"`
	CreateTime time.Time `form:"createTime" time_format:"unixNano"`
	UnixTime   time.Time `form:"unixTime" time_format:"unix"`
}

func startPage(c *gin.Context) {
	var person Person
	if c.ShouldBindQuery(&person) == nil {
		log.Println("====== Only Bind By Query String ======")
		log.Println("Name:", person.Name)
		log.Println("Address:", person.Address)

		c.JSON(http.StatusOK, person)
	} else {
		c.String(http.StatusBadRequest, "invalid parameters")
	}
}

func startPage1(c *gin.Context) {
	var person Person
	// If `GET`, only `Form` binding engine (`query`) used.
	// If `POST`, first checks the `content-type` for `JSON` or `XML`, then uses `Form` (`form-data`).
	// See more at https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L88
	err := c.ShouldBind(&person)
	if err == nil {
		log.Println(person.Name)
		log.Println(person.Address)
		log.Println(person.Birthday)
		log.Println(person.CreateTime)
		log.Println(person.UnixTime)

		c.String(http.StatusOK, "Success")
	} else {
		log.Println(err)
		c.String(http.StatusBadRequest, "invalid parameters")
	}
}

// Binding from JSON
type Login struct {
	User     string `form:"user" json:"user" xml:"user" binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"` // binding:"-"
}

type Book struct {
	CheckIn  time.Time `form:"check_in" binding:"required,bookabledate" time_format:"2006-01-02"`
	CheckOut time.Time `form:"check_out" binding:"required,gtfield=CheckIn" time_format:"2006-01-02"`
}

var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}

func ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Authed pong")
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token != "okay" {
			c.String(http.StatusUnauthorized, "Unauthorized")
		}
	}
}

type redisKVData struct {
	RKey   string `json:"rKey" binding:"required"`
	RValue string `json:"rValue" binding:"required"`
}