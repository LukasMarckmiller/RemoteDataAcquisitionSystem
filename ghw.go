package main

import (
	"fmt"
	"github.com/jaypipes/ghw"
	"strings"
)

type ghwb struct {
}

func printBlockStorageInfo() {
	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v", err)
	}

	fmt.Printf("%v\n", block)
	for _, disk := range block.Disks {
		fmt.Printf(" %v\n", disk)
		for _, part := range disk.Partitions {
			fmt.Printf("  %v\n", part)
		}
	}
}

func getDisksWithoutBootPart() (error, []*ghw.Disk) {
	block, err := ghw.Block()
	if err != nil {
		//fmt.Printf("Error getting block storage info: %v", err)
		return err, nil
	}

	var disksResult []*ghw.Disk
	isBootPartition := false

	for _, disk := range block.Disks {
		for _, part := range disk.Partitions {
			if strings.HasPrefix(part.MountPoint, "/boot/") {
				isBootPartition = true
				break
			}
		}

		if !isBootPartition {
			disksResult = append(disksResult, disk)
		}

		isBootPartition = false
	}

	return err, disksResult

}
