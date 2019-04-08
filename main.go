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
	/*err,disks := getDisksWithoutBootPart()
	if (err != nil){
		fmt.Printf("Err while trying to retriev Block/Disk info", err)
	}
	for _,disk := range disks{
		fmt.Printf(disk.String())
	}
	*/

	router.Run()
	//router.RunTLS(":5443", "/etc/ssl/certs/server.crt","/etc/ssl/private/server.key")
}
