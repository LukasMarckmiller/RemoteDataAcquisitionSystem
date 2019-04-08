package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func showIndexPage(context *gin.Context) {
	context.HTML(

		http.StatusOK,
		"index.html",
		gin.H{
			"title": "Welcome",
		})
}

func getMountedMedia(context *gin.Context) {

	err, disks := getDisksWithoutBootPart()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}
	context.JSON(http.StatusOK, disks)
}

func getMountedMediaById(context *gin.Context) {
	err, disks := getDisksWithoutBootPart()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}

	paramId := context.Params.ByName("id")
	id, err := strconv.Atoi(paramId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, not an integer."})
		return
	}
	if id >= len(disks) {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, index out of bounds."})
		return
	}

	context.JSON(http.StatusOK, disks[id])
}
