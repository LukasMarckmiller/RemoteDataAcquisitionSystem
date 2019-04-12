package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	//Set the router as default
	router = gin.Default()
	//Load static html content
	router.LoadHTMLFiles("web/index.html")
	router.Static("/css", "web/css")
	router.Static("/js", "web/js")

	router.Use(cors.Default())

	// Define Simple test route
	initRoutes()

	/*USE ONLY FOR LAB ENVIRONMENT, BUILD OWN CONFIG FOR PRODUCTIVE BUILD:
	https://github.com/gin-contrib/cors*/

	//Hardware Rec

	/*err,disks := getDisksWithoutBootPart()
	if (err != nil){
		fmt.Printf("Err while trying to retriev Block/Disk info", err)
	}
	for _,disk := range disks{
		fmt.Printf(disk.String())l
	}
	*/

	router.Run(":8000")
	//router.RunTLS(":5443", "/etc/ssl/certs/server.crt","/etc/ssl/private/server.key")
}
