package main

func initRoutes() {
	router.GET("/", showIndexPage)
	router.GET("/media", getMedia)
	router.GET("/media/:id", getMountedMediaById)
	router.POST("/media/transfer", getIsRemoteTransferPossible)

	router.POST("/image", createAndStartImageJob)
	router.GET("/image/:id", getImageJobById)
	router.DELETE("/image/:id", cancelImageJob)
	//Middleware for Stats
	router.GET("/stats", getStatInfo)

	router.GET("/mounted", getMountedMedia)
	router.GET("/mounted/:path", getDiskSpaceStatus)
	//The encoded value gets automatically decoded by gin. Which is bad if your param is a path, then gin decodes it and interprets the value as route
	router.UseRawPath = true
}
