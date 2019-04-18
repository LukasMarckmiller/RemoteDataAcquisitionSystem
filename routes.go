package main

func initRoutes() {
	router.GET("/", showIndexPage)
	router.GET("/media", getMountedMedia)
	router.GET("/media/:id", getMountedMediaById)
	router.POST("/image", createAndStartImageJob)
	router.GET("/image/:id", getImageJobById)
	//Middleware for Stats
	router.GET("/stats", getStatInfo)
}
