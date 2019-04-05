package main

import (
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	//Set the router as default
	router = gin.Default()
	//Load static html content
	router.LoadHTMLGlob("web/templates/*")
	// Define Simple test route
	initRoutes()

	//Hardware Rec
	printBlockStorageInfo()

	router.Run()
	//router.RunTLS(":5443", "/etc/ssl/certs/server.crt","/etc/ssl/private/server.key")
}
