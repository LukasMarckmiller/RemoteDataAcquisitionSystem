package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type ImageJob struct {
	Running bool
	Id      string
	Option  ImageOption
}

type ImageOption struct {
	Type   ImageType
	Target ImageTarget
}

type ImageType int
type ImageTarget int

const (
	DefaultShell    = "sh"
	AquisitionTool  = "dc3dd"
	InputFileArg    = "if="
	InputFileSubDir = "/dev/"
	InputArgs       = "verb=on" // log=/home/lukas/rfa/log" //tilde (~) doesnt get resolved for log param
	OutputFileArgs  = "of=~/rfa/"
	OutputArgs      = "verb=on" //log=/home/lab02/rfa/log Caution with log on output!! after finishing the copy process it seems all output is written to the log file which can take a long time

	Full ImageType = 0
	Part ImageType = 1

	Remote ImageTarget = 0
	Local  ImageTarget = 1
)

var commandIfOutput strings.Builder
var commandOfOutput strings.Builder

func (i *ImageJob) getCachedOutput() (iCache string, oCache string) {
	sI := commandIfOutput.String()
	sO := commandOfOutput.String()
	commandIfOutput.Reset()
	commandOfOutput.Reset()
	return sI, sO
}

//TODO - Implement Progress by using ofs and check if each junk gets reported to stderr, when this is the case compute max chunk len before and then you know how many chunks left
func (i *ImageJob) run(dev string, imgName string) error {
	defer func() { i.Running = false }()

	i.Running = true

	_, err := exec.LookPath(AquisitionTool)
	if err != nil {
		fmt.Println("Installing " + AquisitionTool + "required")
		return err
	}

	r, w := io.Pipe()
	defer w.Close()

	commandIf := fmt.Sprintf("%v %v%v%v %v", AquisitionTool, InputFileArg, InputFileSubDir, dev, InputArgs)
	commandOf := fmt.Sprintf("ssh %v -C '%v %v %v%v'", app.Server, AquisitionTool, OutputArgs, OutputFileArgs, imgName)

	cmdIf := exec.Command(DefaultShell, "-c", commandIf)
	cmdIf.Stdout = w

	cmdOf := exec.Command(DefaultShell, "-c", commandOf)
	cmdOf.Stdin = r

	readerStderrIf, err := cmdIf.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	readerStderrOf, err := cmdOf.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err = cmdIf.Start(); err != nil {
		fmt.Println(err)
		return err
	}

	if err = cmdOf.Start(); err != nil {
		fmt.Println(err)
		return err
	}

	go func() {
		defer fmt.Println("StderrIf done")

		reader := bufio.NewReader(readerStderrIf)

		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					fmt.Println(err)
					break
				}
			}

			m := string(line)
			if m != "" {
				commandIfOutput.WriteString(m)
			}
			//	fmt.Println("StderrIf: " + m)
		}
	}()

	go func() {
		defer fmt.Println("StderrOf done")

		reader := bufio.NewReader(readerStderrOf)

		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					fmt.Println(err)
					break
				}
			}

			m := string(line)
			if m != "" {
				commandOfOutput.WriteString(m)
			}
			//fmt.Println("StderrOf: " + m)
		}
	}()

	fmt.Println("Waiting for Input Command")
	if err := cmdIf.Wait(); err != nil {
		fmt.Println(err)
		return err
	} else {
		fmt.Println("Input Command ended successfully")
	}
	fmt.Println("Writer closing")
	if err := w.Close(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := cmdOf.Wait(); err != nil {
		fmt.Println(err)
		return err
	} else {
		fmt.Println("Output Command ended successfully")
	}
	//At this point either command1 is finished and command2 is finished without errors
	//or command1 is not finished and command2 is finished with errors
	fmt.Println("Done")
	return nil
}
