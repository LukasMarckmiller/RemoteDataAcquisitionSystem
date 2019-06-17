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
	CmdIf   *exec.Cmd
	CmdOf   *exec.Cmd
}

type ImageOption struct {
	Type       ImageType   `json:"type"`
	Target     ImageTarget `json:"target"`
	Compressed bool        `json:"compressed"`
}

type ImageType int
type ImageTarget int

const (
	DefaultShell    = "sh"
	AquisitionTool  = "dc3dd"
	InputFileArg    = "if="
	InputFileSubDir = "/dev/"
	InputArgs       = "verb=on" // log=/home/lukas/rfa/log" //tilde (~) doesnt get resolved for log param
	OutputFileArgs  = "hof=~/rfa/"
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

	commandIf := fmt.Sprintf("%v hash=sha256 hash=md5 hlog=%v.hash %v%v%v %v", AquisitionTool, imgName, InputFileArg, InputFileSubDir, dev, InputArgs) //| gzip -c -1
	var commandOf string

	if i.Option.Target == Local {
		commandOf = fmt.Sprintf("%v %v %v%v.img'", AquisitionTool, OutputArgs, OutputFileArgs, imgName)
	} else {
		commandOf = fmt.Sprintf("ssh %v -C '%v hash=sha256 hash=md5 hlog=%v.hash %v %v%v.img'", app.Server, AquisitionTool, imgName, OutputArgs, OutputFileArgs, imgName)
	}

	i.CmdIf = exec.Command(DefaultShell, "-c", commandIf)
	i.CmdIf.Stdout = w

	i.CmdOf = exec.Command(DefaultShell, "-c", commandOf)
	i.CmdOf.Stdin = r

	readerStderrIf, err := i.CmdIf.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	readerStderrOf, err := i.CmdOf.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err = i.CmdIf.Start(); err != nil {
		fmt.Println(err)
		return err
	}

	if err = i.CmdOf.Start(); err != nil {
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
	if err := i.CmdIf.Wait(); err != nil {
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

	if err := i.CmdOf.Wait(); err != nil {
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

func (i *ImageJob) cancel() error {
	if err := i.CmdIf.Process.Kill(); err != nil {
		return err

	}

	if err := i.CmdOf.Process.Kill(); err != nil {
		return err
	}

	return nil
}
