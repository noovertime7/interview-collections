package main

import "github.com/gin-gonic/gin"

func main() {
	c := gin.New()
	c.Handle("GET", "/user", func(context *gin.Context) {
		context.Writer.Write([]byte("user"))
	})
	c.Handle("GET", "/userlist", func(context *gin.Context) {
		context.Writer.Write([]byte("userlist"))
	})
	c.Run(":9090")
}
