package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	r.GET("/", func(ctx *gin.Context) {
		ctx.Status(200)
		fmt.Fprint(ctx.Writer, "hello")
		ctx.Writer.Flush()
		fmt.Fprint(ctx.Writer, "world")
	})

	r.Run(":8082")
}
