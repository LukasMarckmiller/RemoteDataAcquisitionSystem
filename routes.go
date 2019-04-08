package main

func initRoutes() {
	router.GET("/", showIndexPage)
	router.GET("/media", getMountedMedia)
	router.GET("/media/:id", getMountedMediaById)
}
