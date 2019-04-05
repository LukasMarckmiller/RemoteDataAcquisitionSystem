package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var router *gin.Engine

func main() {
	//Set the router as default
	router = gin.Default()
	//Load static html content
	router.LoadHTMLGlob("web/templates/*")
	// Define Simple test route
	router.GET("/", func(context *gin.Context) {
		context.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"title": "Demo App",
			})
	})

	router.Run()
	//router.RunTLS("127.0.0.1:5443","","")
}
