package gon

import (
	"log"
	"net/http"
)
//简单封装下ResponseWriter，后期可以实现Flusher、Hijacker等接口

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type ResponseWriter interface {
	http.ResponseWriter
	//返回状态码
	Status() int
	//返回body已经写入的数据大小，以bytes作为单位
	Size() int
	//是否已经写入
	Written() bool
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = &responseWriter{}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			log.Printf("override status code %d with %d", w.status, code)
		}
		w.status = code
	}
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}