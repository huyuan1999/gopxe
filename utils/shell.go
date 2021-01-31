package utils

import (
	"io/ioutil"
	"os/exec"
	"syscall"
)

type ShellResult struct {
	Code   int    `json:"code"`
	Pid    int    `json:"pid"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Err    error  `json:"err"`
}

func Shell(command string, args ...string) *ShellResult {
	result := &ShellResult{}

	cmd := exec.Command(command, args...)

	stdoutP, _ := cmd.StdoutPipe()
	stderrP, _ := cmd.StderrPipe()

	defer func() { _ = stdoutP.Close() }()
	defer func() { _ = stderrP.Close() }()

	if err := cmd.Start(); err != nil {
		result.Err = err
		return result
	}

	stdoutResult, err := ioutil.ReadAll(stdoutP)
	if err != nil {
		result.Err = err
		return result
	}

	stderrResult, err := ioutil.ReadAll(stderrP)
	if err != nil {
		result.Err = err
		return result
	}

	if err := cmd.Wait(); err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			result.Code = ex.Sys().(syscall.WaitStatus).ExitStatus()
		}
	}

	result.Pid = cmd.Process.Pid
	result.Stdout = string(stdoutResult)
	result.Stderr = string(stderrResult)
	return result
}
