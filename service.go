package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// 创建一个默认的 Gin 引擎
	r := gin.Default()

	// 定义路由处理程序
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, Gin!")
	})

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
