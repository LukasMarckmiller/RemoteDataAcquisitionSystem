/*
Written by Lukas Marckmiller
This file contains functions for block device handling based on jaypipes/gwh api.
*/
package main

import (
	"fmt"
	"github.com/jaypipes/ghw"
	"golang.org/x/sys/unix"
	"strings"
)

type DiskSpace struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

//Returns available disk space (DiskSpace) for mounted partition (path)
func getAvailableDiskSpace(path string) (disk DiskSpace) {
	fs := unix.Statfs_t{}
	if err := unix.Statfs(path, &fs); err != nil {
		return
	}

	//Convert to bytes
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bavail * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

//Helper function for debugging, currently not used
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

//Returns a list of ghw.Disk objects containing all available block devices.
//The Disk containing the boot partition is filtered.
func getDisksWithoutBootPart() (err error, disks []ghw.Disk) {
	block, err := ghw.Block()
	if err != nil {
		//fmt.Printf("Error getting block storage info: %v", err)
		return
	}

	isBootPartition := false

	for _, disk := range block.Disks {
		for _, part := range disk.Partitions {
			if strings.HasPrefix(part.MountPoint, "/boot") {
				isBootPartition = true
				break
			}
		}

		if !isBootPartition {
			disks = append(disks, *disk)
		}

		isBootPartition = false
	}

	return
}

//Returns a list of ghw.Partition objects containing all partitions for all available block devices where the partition mountpoint is not null.
//The boot Partition is filtered.
func getMountPointsWithoutBoot() (err error, parts []ghw.Partition) {

	//TODO HandleReadOnly Mounts. But dont just hide them because the user needs to know its read-only
	block, err := ghw.Block()
	if err != nil {
		//fmt.Printf("Error getting block storage info: %v", err)
		return
	}

	for _, disk := range block.Disks {
		for _, part := range disk.Partitions {
			if part.MountPoint != "" && !strings.HasPrefix(part.MountPoint, "/boot") {
				parts = append(parts, *part)
			}
		}
	}

	return
}
