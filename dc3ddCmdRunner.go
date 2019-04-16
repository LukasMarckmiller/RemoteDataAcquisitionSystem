package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ImageJob struct {
	Running        bool
	Id             string
	CIf            chan string
	COf            chan string
	CIfcachedValue string
	COfCachedValue string
}

//TODO - Implement Progress by using ofs and check if each junk gets reported to stderr, when this is the case compute max chunk len before and then you know how many chunks left
func (i *ImageJob) runDc3dd(dev string) {
	var commandIfOutput strings.Builder
	var commandOfOutput strings.Builder

	i.Running = true

	path, err := exec.LookPath("dc3dd")
	if err != nil {
		log.Fatal("Installing dc3dd required")
	}

	r, w, err := os.Pipe()
	defer r.Close()

	commandIf := path + " if=/home/lukas/test.txt verb=on log=/home/lukas/dc3ddTest/sdb1Img.log"
	commandOf := "ssh lab01@192.168.0.10 -t \"dc3dd verb=on of=/home/lab01/sdb1/sda1.img \""

	cmdIf := exec.Command("sh", "-c", commandIf)
	cmdIf.Stdout = w

	cmdOf := exec.Command("sh", "-c", commandOf)
	cmdOf.Stdin = r

	readerStderrIf, err := cmdIf.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	readerStderrOf, err := cmdOf.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err = cmdIf.Start(); err != nil {
		log.Fatal(err)
	}

	if err = cmdOf.Start(); err != nil {
		log.Fatal(err)
	}

	//	var wg sync.WaitGroup
	//	wg.Add(1)
	go func() {
		//defer wg.Done()
		defer fmt.Println("StderrIf done")
		defer w.Close()

		scanner := bufio.NewScanner(readerStderrIf)
		for scanner.Scan() {
			m := scanner.Text()
			commandIfOutput.WriteString(m)
			i.CIfcachedValue = commandIfOutput.String()
			fmt.Println("StderrIf: " + m)
		}
	}()

	go func() {
		//		defer wg.Done()
		defer fmt.Println("StderrOf done")

		scanner := bufio.NewScanner(readerStderrOf)
		for scanner.Scan() {
			m := scanner.Text()
			commandOfOutput.WriteString(m)
			i.COfCachedValue = commandOfOutput.String()
			fmt.Println("StderrOf: " + m)
		}
	}()

	if err := cmdOf.Wait(); err != nil {
		fmt.Println(err)
	}

	//At this point either command1 is finished and command2 is finished without errors
	//or command1 is not finished and command2 is finished with errors

	if err := cmdIf.Wait(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Done")
	i.Running = false
}
