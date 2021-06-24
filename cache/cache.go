package cache

import (
	"bytes"
	"golang.org/x/sync/singleflight"
	"gon"
	"gon/cache/persist"
	"log"
	"net/http"
	"sync"
	"time"
)

type Option struct {
	CacheStore persist.CacheStore
	CacheDuration time.Duration
	//是否用singleflight
	UseSingleFlight bool
	//singleflight key的缓存时间
	SingleflightForgetTime time.Duration
}





type KeyGenerator func(c *gon.Context) (string, bool)


func Cache(keyGenerator KeyGenerator, opt Option) gon.HandlerFunc {
	if opt.CacheStore == nil {
		panic("CacheStore can not be nil")
	}

	cacheHelper := newCacheHelper(opt)

	return func(c *gon.Context) {
		cacheKey, needCache := keyGenerator(c)
		if !needCache {
			c.Next()
			return
		}


		//优先读缓存
		{
			//从对象池里拿responseCache
			respCache := cacheHelper.getResponseCache()
			defer cacheHelper.putResponseCache(respCache)

			err := opt.CacheStore.Get(cacheKey, &respCache)
			//读缓存成功
			if err == nil {
				log.Printf("get cache success, cache key: %s", cacheKey)
				cacheHelper.respondWithCache(c, respCache)
				return
			}
			//缓存未命中
			if err == persist.ErrCacheMiss {
				log.Printf("get cache miss, cache key: %s", cacheKey)

			} else {
				log.Printf("get cache error: %s, cache key: %s", err, cacheKey)
			}
		}

		//缓存未命中直接执行路由函数

		//绑定c.Writer和cacheWriter
		cacheWriter := &responseCacheWriter{}
		cacheWriter.reset(c.Writer)
		c.Writer = cacheWriter

		respCache := &responseCache{}

		if !opt.UseSingleFlight {
			c.Next()
		//避免缓存击穿
		} else {
			rawCacheWriter, _, _ := cacheHelper.sfGroup.Do(cacheKey, func() (interface{}, error) {
				if opt.SingleflightForgetTime > 0 {
					go func() {
						time.Sleep(opt.SingleflightForgetTime)
						cacheHelper.sfGroup.Forget(cacheKey)
					}()
				}

				c.Next()

				return cacheWriter, nil
			})
			//接口转成对象
			cacheWriter = rawCacheWriter.(*responseCacheWriter)

		}

		respCache.fill(cacheWriter)





		//更新缓存
		if err := opt.CacheStore.Set(cacheKey, respCache, opt.CacheDuration); err != nil {
			log.Printf("set cache key error: %s, cache key: %s", err, cacheKey)
		}



	}
}

//通过不同方式生成key
//只获取path，不包括querystring等部分
func CacheByPath(opt Option) gon.HandlerFunc {
	return Cache(
		func(c *gon.Context) (string, bool) {
			return c.Request.URL.Path, true
		},
		opt,
	)
}

func CacheByURI(opt Option) gon.HandlerFunc {
	return Cache(
		func(c *gon.Context) (string, bool) {
			return c.Request.RequestURI, true
		},
		opt,
	)
}




type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}
func (c *responseCache) reset() {
	c.Data = c.Data[0:0]
	c.Header = make(http.Header)
}
//用responseCacheWriter辅助写入缓存
func (c *responseCache) fill(cacheWriter *responseCacheWriter) {
	c.Status = cacheWriter.Status()
	c.Data = cacheWriter.body.Bytes()
	c.Header = make(http.Header, len(cacheWriter.Header()))

	for key, value := range cacheWriter.Header() {
		c.Header[key] = value
	}
}




//辅助写缓存
type responseCacheWriter struct {
	gon.ResponseWriter
	body bytes.Buffer
}
//重写Write方法，在服务写入响应body的同时写入到responseCacheWriter中
func (w *responseCacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseCacheWriter) reset(writer gon.ResponseWriter) {
	w.body.Reset()
	w.ResponseWriter = writer
}




type cacheHelper struct {
	sfGroup singleflight.Group
	//增加重用率，减少GC负担
	responseCachePool *sync.Pool
	option Option
}

func newCacheHelper(opt Option) *cacheHelper {
	return &cacheHelper{
		sfGroup:           singleflight.Group{},
		responseCachePool: newResponseCachePool(),
		option:           opt,
	}
}

//获取空响应缓存
func (m *cacheHelper) getResponseCache() *responseCache {
	respCache := m.responseCachePool.Get().(*responseCache)
	respCache.reset()

	return respCache
}

func (m *cacheHelper) putResponseCache(c *responseCache) {
	m.responseCachePool.Put(c)
}

//将缓存写入Response
func (m *cacheHelper) respondWithCache(c *gon.Context, respCache *responseCache) {
	c.Writer.WriteHeader(respCache.Status)
	for k, vals := range respCache.Header {
		for _, v := range vals {
			c.Writer.Header().Set(k, v)
		}
	}

	if _, err := c.Writer.Write(respCache.Data); err != nil {
		log.Printf("write response error: %s", err)
	}

	//return directly
	c.Abort()
}

//生成responseCache
func newResponseCachePool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return &responseCache{
				Header: make(http.Header),
			}
		},
	}
}