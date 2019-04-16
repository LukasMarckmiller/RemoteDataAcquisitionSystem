package main

import (
	"github.com/gin-gonic/gin"
	"github.com/twinj/uuid"
	"net/http"
	"strconv"
)

var jobs = map[string]*ImageJob{}

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

func createAndStartImageJob(context *gin.Context) {
	cOf := make(chan string)
	cIf := make(chan string)
	job := ImageJob{Id: uuid.NewV4().String(), CIf: cIf, COf: cOf}
	go job.runDc3dd(context.PostForm("path"))
	jobs[job.Id] = &job

	context.JSON(http.StatusOK, job.Id)
}

func getImageJobById(context *gin.Context) {
	elem, ok := jobs[context.Params.ByName("id")]
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, no image job for id."})
		return
	}

	context.JSON(http.StatusOK, ImageJobPresentationType{CommandOfOutput: elem.COfCachedValue, CommandIfOutput: elem.CIfcachedValue, Running: elem.Running, Id: elem.Id})
}

//Implement cache cleaning for ImageJobs

type ImageJobPresentationType struct {
	CommandOfOutput string `json:"commandOfOutput"`
	CommandIfOutput string `json:"commandIfOutput"`
	Running         bool   `json:"running"`
	Id              string `json:"id"`
}
