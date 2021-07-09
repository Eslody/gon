# gon
仿写gin实现的golang框架。



在golang原生net/http库静态路由的基础上实现了动态路由，将原本req和resp的参数形式封装成context上下文。其中，路由采用前缀树进行拆分。

实现了静态资源服务机制。

实现了分组控制，支持前缀区分和分组嵌套。

实现了中间件机制，默认实现了日志中间件和错误处理中间件。中间件可作用于子分组。

实现了缓存中间件，对golang官方cache库作了一定程度优化。

在高并发、大数据量写入缓存、高缓存更新频率（2s）的条件下，使用ab压测工具对gon-cache进行多次并发量为1000、时间为一分钟的压测：

```shell
abs -t 60 -c 1000 127.0.0.1:8082/v1/hello
```

最终，gon-cache的qps稳定在1200~1400，如图：

```shell
abs\Apache24\bin>abs -t 60 -c 1000 127.0.0.1:8082/v1/hello
This is ApacheBench, Version 2.3 <$Revision: 1879490 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 5000 requests
Completed 10000 requests
Completed 15000 requests
Completed 20000 requests
Completed 25000 requests
Completed 30000 requests
Completed 35000 requests
Completed 40000 requests
Completed 45000 requests
Completed 50000 requests
Finished 50000 requests


Server Software:
Server Hostname:        127.0.0.1
Server Port:            8082

Document Path:          /v1/hello
Document Length:        20590 bytes

Concurrency Level:      1000
Time taken for tests:   41.087 seconds
Complete requests:      50000
Failed requests:        0
Total transferred:      1033600000 bytes
HTML transferred:       1029500000 bytes
Requests per second:    1216.94 [#/sec] (mean)
Time per request:       821.735 [ms] (mean)
Time per request:       0.822 [ms] (mean, across all concurrent requests)
Transfer rate:          24566.92 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.6      0      48
Processing:     2  791 750.8    404    2355
Waiting:        0  602 747.4    230    2087
Total:          3  792 750.8    404    2355

Percentage of the requests served within a certain time (ms)
  50%    404
  66%    434
  75%   1736
  80%   1964
  90%   2119
  95%   2144
  98%   2189
  99%   2217
 100%   2355 (longest request)
```

而用gin-cache进行相同的测试，结果令人惊讶：

```shell
abs\Apache24\bin>abs -t 60 -c 1000 127.0.0.1:8080/cache_ping
This is ApacheBench, Version 2.3 <$Revision: 1879490 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking 127.0.0.1 (be patient)
Completed 5000 requests
Completed 10000 requests
Completed 15000 requests
Finished 17629 requests


Server Software:
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /cache_ping
Document Length:        20328 bytes

Concurrency Level:      1000
Time taken for tests:   60.019 seconds
Complete requests:      17629
Failed requests:        8452
   (Connect: 0, Receive: 0, Length: 8452, Exceptions: 0)
Total transferred:      42199018688 bytes
HTML transferred:       42197217980 bytes
Requests per second:    293.72 [#/sec] (mean)
Time per request:       3404.565 [ms] (mean)
Time per request:       3.405 [ms] (mean, across all concurrent requests)
Transfer rate:          686614.73 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   3.1      0     241
Processing:    15 2913 4249.0   1224   18821
Waiting:        3  343 419.7    134    1690
Total:         16 2914 4249.3   1225   18822

Percentage of the requests served within a certain time (ms)
  50%   1225
  66%   2070
  75%   3391
  80%   3913
  90%   9868
  95%  14019
  98%  17738
  99%  17928
 100%  18822 (longest request)
```

足足相差了三倍！

当然，这并不是正常条件下的结果，需要较多次的缓存更新写入以及高并发请求，但我们也可以从中取得一些缓存优化的思路。
