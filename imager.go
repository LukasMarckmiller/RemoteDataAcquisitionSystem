//Written by Lukas Marckmiller
//This file controls the image creation and hash validation.
package main

import (
	"bufio"
	"fmt"
	"github.com/jaypipes/ghw"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

type ImageJob struct {
	Running    bool        `json:"running"`
	Id         string      `json:"id"`
	Option     ImageOption `json:"option"`
	CmdIf      *exec.Cmd   `json:"cmd_if"`
	CmdOf      *exec.Cmd   `json:"cmd_of"`
	Hashes     Hashes      `json:"hashes"`
	HashResult HashResult  `json:"hash_result"`
}

type ImageOption struct {
	Type       ImageType   `json:"type"`
	Target     ImageTarget `json:"target"`
	Compressed bool        `json:"compressed"`
}

type ImageType int
type ImageTarget int

const (
	DefaultExtension = ".img"
	DefaultShell     = "sh"

	//DC3DD
	AquisitionTool  = "dc3dd"
	InputFileArg    = "if="
	InputFileSubDir = "/dev/"
	InputArgs       = "verb=on" // log=/home/lukas/rfa/log" //tilde (~) doesnt get resolved for log param
	OutputFileArgs  = "hof="
	OutputArgs      = "verb=on" //log=/home/lab02/rfa/log Caution with log on output!! after finishing the copy process it seems all output is written to the log file which can take a long time

	Full ImageType = 0
	Part ImageType = 1

	Remote ImageTarget = 0
	Local  ImageTarget = 1
)

var (
	md5Regex    = regexp.MustCompile(`[a-f0-9]{32} \(md5\)`)
	sha256Regex = regexp.MustCompile(`[A-Fa-f0-9]{64} \(sha256\)`)

	//Used for progress reports. Contains only one line per call.
	commandIfOutput strings.Builder
	commandOfOutput strings.Builder

	//Used for logging. Contains complete Output
	bufferedOutput strings.Builder
	bufferedInput  strings.Builder
	pipeWriter     *io.PipeWriter
	pipeReader     *io.PipeReader
)

type Hashes struct {
	Md5Input  string `json:"md_5_input"`
	Md5Output string `json:"md_5_output"`

	Sha256Input  string `json:"sha_256_input"`
	Sha256Output string `json:"sha_256_output"`
}

type HashResult struct {
	Md5Valid    bool `json:"md_5_valid"`
	Sha256Valid bool `json:"sha_256_valid"`
}

//Returns incremental cachedOutput as string from input and output command.
func (i *ImageJob) getCachedOutput() (string, string) {
	sI := commandIfOutput.String()
	sO := commandOfOutput.String()
	commandIfOutput.Reset()
	commandOfOutput.Reset()
	return sI, sO
}

//TODO - Implement Progress by using ofs and check if each junk gets reported to stderr, when this is the case compute max chunk len before and then you know how many chunks left
//Runs and creates the image with preconfigured tools and settings for a dev.
//The mountTarget is only used if ImageJob.imageOptions.target is local.
//Output of the commands is written to a buffer which can be obtained by calling getCachedOutput() for only the newest output or using buffered... vars.
func (i *ImageJob) run(dev string, mountTarget ghw.Partition, imgName string) error {
	defer func() { i.Running = false }()

	i.Running = true

	_, err := exec.LookPath(AquisitionTool)
	if err != nil {
		fmt.Println("Installing " + AquisitionTool + "required")
		return err
	}

	pipeReader, pipeWriter = io.Pipe()
	defer pipeWriter.Close()

	extension := DefaultExtension

	var commandIf string
	var commandOf string

	//Set defaults
	commandIf = fmt.Sprintf("%s hash=sha256 hash=md5 hlog=%s.hash %s %s%s%s", AquisitionTool, imgName, InputArgs, InputFileArg, InputFileSubDir, dev)

	//Automatically compress/decompress transmission.
	if i.Option.Target == Remote {
		commandIf += " | gzip -c -1"
		if i.Option.Compressed {
			extension += ".gz"
			commandOf = fmt.Sprintf("ssh %s -C '%s hash=sha256 hash=md5 hlog=%s.hash %s %s%s%s'", app.Server, AquisitionTool, imgName, OutputArgs, OutputFileArgs, imgName, extension)
		} else {
			commandOf = fmt.Sprintf("ssh %s -C 'gzip -d | %s hash=sha256 hash=md5 hlog=%s.hash %s %s%s%s'", app.Server, AquisitionTool, imgName, OutputArgs, OutputFileArgs, imgName, extension)
		}
	} else {
		//Compress local image if option is set
		if i.Option.Compressed {
			commandIf += " | gzip -c -1"
			extension += ".gz"
		}
		commandOf = fmt.Sprintf("%s hash=sha256 hash=md5 hlog=%s.hash  %s %s%s%s", AquisitionTool, imgName, OutputArgs, OutputFileArgs, mountTarget.MountPoint+"/"+imgName, extension)
	}

	fmt.Println(commandIf, commandOf)
	i.CmdIf = exec.Command(DefaultShell, "-c", commandIf)
	i.CmdIf.Stdout = pipeWriter

	i.CmdOf = exec.Command(DefaultShell, "-c", commandOf)
	i.CmdOf.Stdin = pipeReader

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
				bufferedInput.WriteString(m)
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
				bufferedOutput.WriteString(m)
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
	if err := pipeWriter.Close(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := i.CmdOf.Wait(); err != nil {
		fmt.Println(err)
		return err
	} else {
		fmt.Println("Output Command ended successfully")
	}

	//TODO Transfer Hash Log to Server
	i.HashResult = i.verfiyHashes()
	bufferedOutput.Reset()
	bufferedInput.Reset()
	fmt.Println("Done")
	return nil
}

//Tries to close the associated pipe and to kill the input and output process
//Returns an error if not successful
func (i *ImageJob) cancel() error {
	_ = pipeWriter.Close()
	_ = pipeReader.Close()
	if err := i.CmdIf.Process.Kill(); err != nil {
		return err

	}

	if err := i.CmdOf.Process.Kill(); err != nil {
		return err
	}

	return nil
}

//Is called after Image is successfully created.
//Checks if both hashes are equal for the input and the created image.
func (i *ImageJob) verfiyHashes() (hashResult HashResult) {
	i.Hashes.Md5Input = md5Regex.FindString(bufferedInput.String())
	i.Hashes.Sha256Input = sha256Regex.FindString(bufferedInput.String())
	i.Hashes.Md5Output = md5Regex.FindString(bufferedOutput.String())
	i.Hashes.Sha256Output = sha256Regex.FindString(bufferedOutput.String())

	if i.Hashes.Md5Output != "" && i.Hashes.Md5Output == i.Hashes.Md5Input {
		hashResult.Md5Valid = true
	}
	if i.Hashes.Sha256Output != "" && i.Hashes.Sha256Input == i.Hashes.Sha256Output {
		hashResult.Sha256Valid = true
	}
	return hashResult
}
