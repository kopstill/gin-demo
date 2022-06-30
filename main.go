package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v9"

	"kopever/gin-demo/testdata/protoexample"
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

	// Multipart/Urlencoded binding
	router.POST("/profile", profileHandler)

	// XML, JSON, YAML and ProtoBuf rendering
	router.GET("/someJSON", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})
	router.GET("/moreJSON", func(c *gin.Context) {
		var msg struct {
			Name    string `json:"user"`
			Message string
			Number  int
		}
		msg.Name = "Lena"
		msg.Message = "hey"
		msg.Number = 123
		c.JSON(http.StatusOK, msg)
	})
	router.GET("/someXML", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})
	router.GET("/someYAML", func(c *gin.Context) {
		c.YAML(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})
	router.GET("/someProtoBuf", func(c *gin.Context) {
		reps := []int64{int64(1), int64(2)}
		label := "test"
		data := &protoexample.Test{
			Label: &label,
			Reps:  reps,
		}
		c.ProtoBuf(http.StatusOK, data)
	})

	// SecureJSON
	// router.SecureJsonPrefix(")]}',\n")
	router.GET("/someJSONSecure", func(c *gin.Context) {
		names := []string{"lena", "austin", "foo"}
		c.SecureJSON(http.StatusOK, names)
	})

	// JSONP
	router.GET("/JSONP", func(c *gin.Context) {
		data := gin.H{
			"foo": "bar",
		}
		c.JSONP(http.StatusOK, data)
		// curl http://127.0.0.1:8080/JSONP?callback=x
	})

	// AsciiJSON
	router.GET("/someJSONAscii", func(c *gin.Context) {
		data := gin.H{
			"lang": "GO 语言",
			"tag":  "<br>",
		}
		c.AsciiJSON(http.StatusOK, data)
	})

	// PureJSON
	router.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"html": "<b>Hello, world!</b>",
		})
	})
	router.GET("/purejson", func(c *gin.Context) {
		c.PureJSON(http.StatusOK, gin.H{
			"html": "<b>Hello, world!</b>",
		})
	})

	// Serving static files
	router.Static("/assets", "./assets")
	router.StaticFS("/more_static", http.Dir("my_file_system"))
	router.StaticFile("/favicon.ico", "./resources/favicon.svg")
	router.StaticFileFS("/more_favicon.ico", "ok.png", http.Dir("my_file_system"))

	// Serving data from file
	router.GET("/local/file", func(c *gin.Context) {
		c.File("local/hello.go")
	})
	var fs http.FileSystem = http.Dir(".")
	router.GET("/fs/file", func(c *gin.Context) {
		c.FileFromFS("local/world.go", fs)
	})

	// Serving data from reader
	router.GET("/someDataFromReader", func(c *gin.Context) {
		response, err := http.Get("https://raw.githubusercontent.com/gin-gonic/logo/master/color.png")
		if err != nil || response.StatusCode != http.StatusOK {
			c.Status(http.StatusServiceUnavailable)
			return
		}

		reader := response.Body
		defer reader.Close()
		contentLength := response.ContentLength
		contentType := response.Header.Get("Content-Type")

		extraHeaders := map[string]string{
			"Content-Disposition": `attachment; filename="gopher.png"`,
		}

		c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
	})

	// HTML rendering
	// router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	// router.LoadHTMLFiles("templates/index.tmpl")
	router.LoadHTMLGlob("templates/*.tmpl")
	// tmpl := template.Must(template.ParseFiles("templates/index.tmpl"))
	// router.SetHTMLTemplate(tmpl)
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})
	router.LoadHTMLGlob("templates/**/*")
	router.GET("/posts/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "posts/index.tmpl", gin.H{
			"title": "Posts",
		})
	})
	router.GET("/users/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/index.tmpl", gin.H{
			"title": "Users",
		})
	})

	// Custom Template renderer
	html := template.Must(template.ParseFiles("templates/template1.tmpl", "templates/template2.tmpl"))
	router.SetHTMLTemplate(html)
	// Custom Delimiters
	router.Delims("{[{", "}]}")
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})
	// Custom Template Funcs
	router.LoadHTMLFiles("testdata/template/raw.tmpl")
	router.GET("/raw", func(c *gin.Context) {
		c.HTML(http.StatusOK, "raw.tmpl", gin.H{
			"now": time.Date(2017, 07, 01, 0, 0, 0, 0, time.UTC),
		})
	})

	// Multitemplate
	// Redirects
	router.GET("/test", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
	})
	// Redirect from POST
	router.POST("/testPost", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/foo")
	})
	// Router redirect
	router.GET("/test1", func(c *gin.Context) {
		c.Request.URL.Path = "/test2"
		router.HandleContext(c)
	})
	router.GET("/test2", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world"})
	})

	// Custom Middleware
	router.Use(Logger())
	router.GET("/customMiddleware", func(c *gin.Context) {
		example := c.MustGet("example").(string)
		log.Print(example)
	})

	// Using BasicAuth() middleware
	adminAuthorized := router.Group("/admin", gin.BasicAuth(gin.Accounts{
		"foo":    "bar",
		"austin": "1234",
		"lena":   "hello2",
		"manu":   "4321",
	}))
	adminAuthorized.GET("/secrets", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		if secret, ok := secrets[user]; ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": "NO SECRET :("})
		}
	})

	// Goroutines inside a middleware
	router.GET("/long_async", func(c *gin.Context) {
		cCp := c.Copy()

		go func() {
			time.Sleep(3 * time.Second)
			log.Println("Done! in path " + cCp.Request.URL.Path)
		}()

		c.String(http.StatusOK, "Done!")
	})
	router.GET("/long_sync", func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		log.Println("Done! in path " + c.Request.URL.Path)
	})

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

	// router.Run()
	// Custom HTTP configuration
	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

var secrets = gin.H{
	"foo":    gin.H{"email": "foo@bar.com", "phone": "123433"},
	"austin": gin.H{"email": "austin@example.com", "phone": "666"},
	"lena":   gin.H{"email": "lena@guapa.com", "phone": "523443"},
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		log.Println("start:", t)
		c.Set("example", "12138")
		c.Next()
		latency := time.Since(t)
		log.Println("takes:", latency)
		status := c.Writer.Status()
		log.Println("response status:", status)
	}
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
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
