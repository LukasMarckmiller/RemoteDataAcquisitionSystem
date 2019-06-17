package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jaypipes/ghw"
	"github.com/semihalev/gin-stats"
	"github.com/twinj/uuid"
	"net/http"
	"net/url"
	"strconv"
)

var jobs = map[string]*ImageJob{}
var imageJobError error

func showIndexPage(context *gin.Context) {
	context.HTML(

		http.StatusOK,
		"index.html",
		gin.H{
			"title": "Welcome",
		})
}

func getIsRemoteTransferPossible(context *gin.Context) {
	var device DevicePresentationType
	var cachedOptions ImageOption

	if err := context.BindJSON(&device); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusBadRequest, "message": "Bad request format."})
		return
	}
	estimatedTime, err := getEstimatedTimeInSecs(device.Size, device.Name)
	if err != nil {
		cachedOptions.Target = Local
	} else {

		cachedOptions.Target = Remote
		fullImageTransfer := validateTime(estimatedTime)
		/*
			time := estimatedTime
			h := time / 60 / 60
			time -= h * 60 * 60
			m := time / 60

			fmt.Printf("Estimated time %02d:%02d\n", h, m)
		*/
		if fullImageTransfer {
			cachedOptions.Type = Full
		} else {
			cachedOptions.Type = Part
		}
	}
	context.JSON(http.StatusOK, &ImageOptionsPresentationType{cachedOptions, estimatedTime})
}

func getMedia(context *gin.Context) {

	err, disks := getDisksWithoutBootPart()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}
	context.JSON(http.StatusOK, disks)
}

func getDiskSpaceStatus(context *gin.Context) {
	path := context.Params.ByName("path")
	path, _ = url.QueryUnescape(path)
	path, _ = strconv.Unquote(path)
	context.JSON(http.StatusOK, getAvailableDiskSpace(path))
}

func getMountedMedia(context *gin.Context) {
	err, parts := getMountPointsWithoutBoot()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}
	context.JSON(http.StatusOK, parts)
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
	//Check disk write estimated time and set to ImageJobOptions -> part if low writetime and full if good write time
	var imageJobRequestPresentation ImageJobRequestPresentationType

	if err := context.BindJSON(&imageJobRequestPresentation); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for imageJob."})
		return
	}

	devPath := imageJobRequestPresentation.Path
	cachedOptions := imageJobRequestPresentation.ImageOption
	uuidV4 := uuid.NewV4().String()
	job := ImageJob{Id: uuidV4, Option: cachedOptions}

	go func() { imageJobError = job.run(devPath, "sdbtest.img") }()
	jobs[job.Id] = &job
	context.JSON(http.StatusOK, job.Id)
}

func cancelImageJob(context *gin.Context) {
	elem, ok := jobs[context.Param("id")]
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, no image job for id."})
		return
	}

	if err := elem.cancel(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error while canceling image job."})
		return
	}

	delete(jobs, context.Param("id"))
	context.Status(http.StatusOK)
	return
}

func getImageJobById(context *gin.Context) {
	elem, ok := jobs[context.Param("id")]
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, no image job for id."})
		return
	}
	err := false
	if imageJobError != nil {
		err = true
	}

	inputFileOut, outputFileOut := elem.getCachedOutput()
	context.JSON(http.StatusOK, ImageJobPresentationType{CommandOfOutput: outputFileOut, CommandIfOutput: inputFileOut, Running: elem.Running, Id: elem.Id, Error: err})
}

func getStatInfo(context *gin.Context) {
	context.JSON(http.StatusOK, stats.Report())
}

//TODO Implement cache cleaning for ImageJobs

type ImageJobPresentationType struct {
	CommandOfOutput string `json:"commandOfOutput"`
	CommandIfOutput string `json:"commandIfOutput"`
	Running         bool   `json:"running"`
	Id              string `json:"id"`
	Error           bool
}

type ImageJobRequestPresentationType struct {
	Path        string        `json:"path"`
	ImageOption ImageOption   `json:"image_option"`
	Mount       ghw.Partition `json:"mount"`
}

type DevicePresentationType struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ImageOptionsPresentationType struct {
	ImageOption   ImageOption `json:"image_option"`
	EstimatedSecs int32       `json:"estimated_secs"`
}
