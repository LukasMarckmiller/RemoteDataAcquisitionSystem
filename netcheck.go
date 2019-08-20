//Written by Lukas Marckmiller
//This file contains helper functions for network bandwidth check and transmission timings ...
package main

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	C            = 1
	K            = C * 1024 //Kilo
	M            = K * 1024 //Mega
	G            = M * 1024 //Giga
	Limit        = 8 * 60   //12 Hours in mins
	TestByteSize = 1 * K
	TestCount    = 10 * K //Transmission total = TestByteSize * TestCount
)

//Returns true if expected transmission time is lower than the limit or false if exceeds the limit.
//timeInSecs is the expected transmission time returned by netcheck(...)
func validate(timeInSecs int32) bool {
	timeInM := timeInSecs / 60
	if timeInM <= Limit {
		return true
	} else {
		return false
	}
}

//Returns expected transmission time for a device (deviceName) with a give filesize.
func netcheck(filesize int64, deviceName string) (int32, error) {
	throughput, err := calcThroughput(deviceName)
	if err != nil {
		return 0, err
	}

	return int32(float64(filesize) / throughput), nil
}

//Calculates throughput as float64 for a give device (deviceName).
//Uses dd for a probe, compressed transmission to the server in app.Server.
//Uses SSH to ensure confidentiality.
//Throughput is computed by measuring the time for executing the dd command with a timeout of 10 Seconds.
func calcThroughput(deviceName string) (throughput float64, err error) {
	//10 MB
	cmdctn := fmt.Sprintf("dd if=/dev/%s bs=%d count=%d | gzip | ssh -C  %v 'gzip -d | cat > /dev/null'", deviceName, TestByteSize, TestCount, app.Server)
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
	//Just to be sure
	timeout = nil

	after := time.Now()
	duration := after.Sub(before).Seconds()

	return (TestByteSize * TestCount) / duration, nil //in byte/s

}
