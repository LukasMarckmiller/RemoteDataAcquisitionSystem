//Written by Lukas Marckmiller
//This file contains the handler funcs for defined rest routes.
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jaypipes/ghw"
	"github.com/semihalev/gin-stats"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//Cache for all active jobs
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

//Handler for network bandwidth check, decides which imager is used for transmission and which output location
func getIsRemoteTransferPossible(context *gin.Context) {
	var device DevicePresentation
	var cachedOptions ImageOption

	if err := context.BindJSON(&device); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusBadRequest, "message": "Bad request format."})
		return
	}
	estimatedTime, err := netcheck(device.Size, device.Name)
	if err != nil {
		cachedOptions.Target = Local
	} else {

		cachedOptions.Target = Remote
		fullImageTransfer := validate(estimatedTime)
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
	context.JSON(http.StatusOK, &ImageOptionsPresentation{cachedOptions, estimatedTime})
}

//Handler returns all plugged in block devices
func getMedia(context *gin.Context) {

	err, disks := getDisksWithoutBootPart()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}
	context.JSON(http.StatusOK, disks)
}

//Handler returns free,used bytes for mountpoint
func getDiskSpaceStatus(context *gin.Context) {
	path := context.Params.ByName("path")
	path, _ = url.QueryUnescape(path)
	path, _ = strconv.Unquote(path)
	context.JSON(http.StatusOK, getAvailableDiskSpace(path))
}

//Handler returns a list of mounted gwh.Partitions
func getMountedMedia(context *gin.Context) {
	err, parts := getMountPointsWithoutBoot()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": "Error retrieving device information."})
		return
	}
	context.JSON(http.StatusOK, parts)
}

//Handler returns mounted ghw.Partition with id
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

//Handler starts imaging process and verify hashes
func createAndStartImageJob(context *gin.Context) {
	//Check disk write estimated time and set to ImageJobOptions -> part if low writetime and full if good write time
	var imageJobRequestPresentation ImageJobRequestPresentation

	if err := context.BindJSON(&imageJobRequestPresentation); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for imageJob."})
		return
	}

	devPath := imageJobRequestPresentation.Path
	cachedOptions := imageJobRequestPresentation.ImageOption
	mountTarget := imageJobRequestPresentation.Mount
	imgName := app.DeviceName + time.Now().Format("20060102MST030405PM")
	job := ImageJob{Id: imgName, Option: cachedOptions}

	go func() {
		imageJobError = job.run(devPath, mountTarget, imgName)
	}()
	jobs[job.Id] = &job
	context.JSON(http.StatusOK, job.Id)
}

//Handler aborts image process with id
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

//Handler returns image job which contains progress information and stats.
func getImageJobById(context *gin.Context) {
	elem, ok := jobs[context.Param("id")]
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Bad input value for parameter id, no image job for id."})
		return
	}
	var imageJobErrorText string
	if imageJobError != nil {
		imageJobErrorText = imageJobError.Error()
	}

	inputFileOut, outputFileOut := elem.getCachedOutput()
	context.JSON(http.StatusOK, ImageJobPresentation{
		CommandOfOutput: outputFileOut,
		CommandIfOutput: inputFileOut,
		Running:         elem.Running,
		Id:              elem.Id,
		Error:           imageJobErrorText,
		Hashes:          elem.Hashes,
		HashResult:      elem.HashResult})
}

//Handler for stats middleware.
func getStatInfo(context *gin.Context) {
	context.JSON(http.StatusOK, stats.Report())
}

//TODO Implement cache cleaning for ImageJobs

type ImageJobPresentation struct {
	CommandOfOutput string     `json:"commandOfOutput"`
	CommandIfOutput string     `json:"commandIfOutput"`
	Running         bool       `json:"running"`
	Id              string     `json:"id"`
	Error           string     `json:"error"`
	Hashes          Hashes     `json:"hashes"`
	HashResult      HashResult `json:"hash_result"`
}

type ImageJobRequestPresentation struct {
	Path        string        `json:"path"`
	ImageOption ImageOption   `json:"image_option"`
	Mount       ghw.Partition `json:"mount"`
}

type DevicePresentation struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ImageOptionsPresentation struct {
	ImageOption   ImageOption `json:"image_option"`
	EstimatedSecs int32       `json:"estimated_secs"`
}
