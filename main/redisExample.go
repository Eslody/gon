package main

import (
	"github.com/go-redis/redis/v8"
	"gon"
	"gon/cache"
	"gon/cache/persist"
	"time"
)


func main() {
	app := gon.Default()

	redisStore := persist.NewRedisStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	}))

	app.Use(cache.CacheByPath(cache.Option{
		CacheDuration: 5 * time.Second,
		CacheStore:    redisStore,
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
