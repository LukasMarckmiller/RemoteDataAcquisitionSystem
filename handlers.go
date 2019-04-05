package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func showIndexPage(context *gin.Context) {
	context.HTML(

		http.StatusOK,
		"index.html",
		gin.H{
			"title": "Demo App",
		})
}
