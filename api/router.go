package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/block-identity/block-identity-server/api/handler"
)

func GetRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "welcome to block identity server",
		})
	})
	r.GET("/auth", handler.Auth)
	r.POST("/transaction", handler.Transaction)
	r.POST("/test", testAuth)
	return r
}

func testAuth(c *gin.Context) {
	fmt.Println("credential パラメータを表示します。")
	fmt.Println(c.GetQuery("credential"))
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
