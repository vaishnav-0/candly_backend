package utils

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func PrintGos() {
	for {
		fmt.Printf("no of gos: %d\n", runtime.NumGoroutine())
		time.Sleep(time.Second)
	}
}


//wrap gin.HandlerFunc to http.Handler
type HttpHandler struct {
	c *gin.Context
	h gin.HandlerFunc
}

func (handler HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.h(handler.c)
}

func WrapGin(handle gin.HandlerFunc, c *gin.Context) http.Handler {
	return HttpHandler{
		c: c,
		h: handle,
	}
}

