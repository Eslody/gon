package main

import (
	"gon"
	"gon/cache"
	"gon/cache/persist"
	"time"
)


func main() {
	app := gon.Default()


	app.Use(cache.CacheByPath(cache.Option{
		CacheDuration: 5 * time.Second,
		CacheStore:          persist.NewMemoryStore(1 * time.Minute),
		UseSingleFlight: true,
	}))
	app.GET("/hello",
		func(c *gon.Context) {
			c.String(200, "hello world")
		},
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
