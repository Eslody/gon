package cache

import (
	"github.com/stretchr/testify/assert"
	"gon"
	"gon/cache/persist"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestCachePath(t *testing.T) {
	cachePathMiddleware := CacheByPath(Option{
		CacheDuration:       5 * time.Second,
		CacheStore:          persist.NewMemoryStore(1 * time.Minute),
		UseSingleFlight: true,
	})

	r := gon.Default()
	r.Use(cachePathMiddleware)

	testBody := "hello world"
	r.GET("/hello", func(c *gon.Context) {
		c.String(200, testBody)
	})

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testWriter := httptest.NewRecorder()
			r.ServeHTTP(testWriter, &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "/hello",
				},
			})

			body, err := ioutil.ReadAll(testWriter.Result().Body)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, string(body), testBody)
		}()
	}

	wg.Wait()
}
