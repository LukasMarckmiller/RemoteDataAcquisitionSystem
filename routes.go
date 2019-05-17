package main

func initRoutes() {
	router.GET("/", showIndexPage)
	router.GET("/media", getMedia)
	router.GET("/media/:id", getMountedMediaById)
	router.POST("/image", createAndStartImageJob)
	router.GET("/image/:id", getImageJobById)
	//Middleware for Stats
	router.GET("/stats", getStatInfo)
	router.GET("/mounted", getMountedMedia)
	router.GET("/mounted/:path", getDiskSpaceStatus)
	router.POST("/media/transfer", getIsRemoteTransferPossible)
	//The encoded value gets automatically decoded by gin. Which is bad if youre param is a path, then gin decodes it and interprets the value as route
	router.UseRawPath = true
}
