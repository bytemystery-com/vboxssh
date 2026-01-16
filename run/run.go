// Package for runnning commands and get te output in a string slice
package run

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

func mergeOutAndErr(bOut bytes.Buffer, bErr bytes.Buffer) []string {
	str := bOut.String()
	if bErr.Len() > 0 {
		if len(str) > 0 {
			str += "\n"
		}
		str += bErr.String()
	}
	lines := strings.Split(str, "\n")
	return lines
}

// runs a command local
func RunLocalCmd(cmd string, args []string, userWriterOut, userWriterErr io.Writer) ([]string, error) {
	if userWriterOut == nil && userWriterErr == nil {
		return RunLocalCmdSimple(cmd, args)
	}
	return RunLocalCmdWithProgess(cmd, args, userWriterOut, userWriterErr)
}

func RunLocalCmdSimple(cmd string, args []string) ([]string, error) {
	cmdEx := exec.Command(cmd, args...)
	var bOut bytes.Buffer
	var bErr bytes.Buffer
	cmdEx.Stdout = &bOut
	cmdEx.Stderr = &bErr
	err := cmdEx.Run()
	lines := mergeOutAndErr(bOut, bErr)
	return lines, err
}

func RunLocalCmdWithProgess(cmd string, args []string, userWriterOut, userWriterErr io.Writer) ([]string, error) {
	cmdEx := exec.Command(cmd, args...)
	var bOut bytes.Buffer
	var bErr bytes.Buffer

	stdout, err := cmdEx.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmdEx.StderrPipe()
	if err != nil {
		return nil, err
	}

	outWriters := make([]io.Writer, 0, 2)
	errWriters := make([]io.Writer, 0, 2)

	outWriters = append(outWriters, &bOut)
	if userWriterOut != nil {
		outWriters = append(outWriters, userWriterOut)
	}

	if userWriterErr != nil {
		errWriters = append(errWriters, userWriterErr)
	}

	mOut := io.MultiWriter(outWriters...)
	mErr := io.MultiWriter(errWriters...)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(mOut, stdout)
	}()

	go func() {
		defer wg.Done()
		io.Copy(mErr, stderr)
	}()

	err = cmdEx.Start()
	if err != nil {
		return nil, err
	}
	err = cmdEx.Wait()
	wg.Wait()

	lines := mergeOutAndErr(bOut, bErr)
	return lines, err
}

// runs a command via SSH
func RunSshCmd(client *ssh.Client, cmd string, args []string, userWriterOut, userWriterErr io.Writer) ([]string, error) {
	if userWriterOut == nil && userWriterErr == nil {
		return RunSshCmdSimple(client, cmd, args)
	}
	return RunSshCmdWithProgress(client, cmd, args, userWriterOut, userWriterErr)
}

func RunSshCmdSimple(client *ssh.Client, cmd string, args []string) ([]string, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var bOut bytes.Buffer
	var bErr bytes.Buffer
	session.Stdout = &bOut
	session.Stderr = &bErr

	for _, arg := range args {
		cmd += " " + arg
	}
	err = session.Run(cmd)
	lines := mergeOutAndErr(bOut, bErr)

	return lines, err
}

func RunSshCmdWithProgress(client *ssh.Client, cmd string, args []string, userWriterOut, userWriterErr io.Writer) ([]string, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	var bOut bytes.Buffer
	var bErr bytes.Buffer

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		cmd += " " + arg
	}

	outWriters := make([]io.Writer, 0, 2)
	errWriters := make([]io.Writer, 0, 2)

	outWriters = append(outWriters, &bOut)
	if userWriterOut != nil {
		outWriters = append(outWriters, userWriterOut)
	}

	errWriters = append(errWriters, &bErr)
	if userWriterErr != nil {
		errWriters = append(errWriters, userWriterErr)
	}

	mOut := io.MultiWriter(outWriters...)
	mErr := io.MultiWriter(errWriters...)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(mOut, stdout)
	}()

	go func() {
		defer wg.Done()
		io.Copy(mErr, stderr)
	}()

	err = session.Start(cmd)
	if err != nil {
		return nil, err
	}
	err = session.Wait()
	wg.Wait()

	lines := mergeOutAndErr(bOut, bErr)
	return lines, err
}
