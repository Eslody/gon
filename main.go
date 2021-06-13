package main

import (
	"gon"
	"log"
	"net/http"
	"time"
)


//一个简单的中间件
func onlyForV2() gon.HandlerFunc {
	return func(c *gon.Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := gon.Default()
	r.GET("/", func(c *gon.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	r.GET("/panic", func(c *gon.Context) {
		nums := []string{"gon"}
		c.String(http.StatusOK, nums[10])
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/hello/:name/*like", func(c *gon.Context) {
			c.String(http.StatusOK, "hello %s, you like %s\n", c.Params["name"], c.Params["like"])
		})
	}

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) //v2组的中间件，是否使用会带来完全不同的结果
	{
		v2.GET("/hello/:name", func(c *gon.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Params["name"], c.Path)
		})
		v2.POST("/login", func(c *gon.Context) {
			c.JSON(http.StatusOK, gon.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}
	r.Static("/assets", "./static")
	r.Run(":9998")
}