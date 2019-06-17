package main

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	C            = 1
	K            = C * 1024
	M            = K * 1024
	G            = M * 1024
	MinBoundMin  = 8 * 60 //12 Hours in mins
	TestByteSize = 1 * K
	TestCount    = 10 * K //Transmission total = TestByteSize * TestCount
)

func validateTime(timeInSecs int32) bool {
	timeInM := timeInSecs / 60
	if timeInM <= MinBoundMin {
		return true
	} else {
		return false
	}
}

func getEstimatedTimeInSecs(filesize int64, deviceName string) (int32, error) {
	throughput, err := getThroughputInBPerSec(deviceName)
	if err != nil {
		return 0, err
	}

	return int32(float64(filesize) / throughput), nil
}

/*Speedtest*/

func getThroughputInBPerSec(deviceName string) (throughput float64, err error) {
	//10 MB
	cmdctn := fmt.Sprintf("dd if=/dev/%s bs=%d count=%d | gzip | ssh -C  %v 'cat > /dev/null'", deviceName, TestByteSize, TestCount, app.Server)
	cmd := exec.Command("sh", "-c", cmdctn)

	timeout := time.AfterFunc(10*time.Second, func() {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Printf("Cant kill network check process.\n%s\n", err)
			return
		}
		fmt.Println("Network check process killed.")
	})

	before := time.Now()
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	timeout.Stop()
	//Stop timeout if finished
	fmt.Println("Timer stopped.")
	//Just to be sure
	timeout = nil

	after := time.Now()
	duration := after.Sub(before).Seconds()

	return (TestByteSize * TestCount) / duration, nil //in byte/s

}
