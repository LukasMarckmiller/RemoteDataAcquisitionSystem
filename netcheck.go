package main

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	C        = 1
	K        = C * 1024
	M        = K * 1024
	G        = M * 1024
	minBound = 12 //12 Hours
)

func validateTime(timeInSecs uint32) bool {
	timeInH := timeInSecs / 60 / 60
	if timeInH <= minBound {
		return true
	} else {
		return false
	}
}

func getEstimatedTimeInSecs(filesize int) (uint32, error) {
	throughput, err := getThroughputInMBPerSec()
	if err != nil {
		return 0, err
	}

	return uint32(float64(filesize) / throughput), nil
}

/*Speedtest*/

func getThroughputInMBPerSec() (throughput float64, err error) {
	//10 MB
	cmdctn := fmt.Sprintf("dd if=/dev/zero bs=%d count=10000 | ssh %v 'cat > /dev/null'", 1*K, app.Server)
	cmd := exec.Command("sh", "-c", cmdctn)
	before := time.Now()
	if err := cmd.Run(); err != nil {
		return -1, err
	}
	after := time.Now()
	duration := after.Sub(before).Seconds()

	return 10 * M / duration, nil //in byte/s

}
