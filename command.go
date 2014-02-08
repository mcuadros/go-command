package command

import (
	"bytes"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

var errorTimeout = errors.New("error: execution timeout")

type ExecutionResponse struct {
	Elapsed  time.Duration
	Stdout   []byte
	Stderr   []byte
	Pid      int
	ExitCode int
	Failed   bool
}

type Command struct {
	response  *ExecutionResponse
	cmd       *exec.Cmd
	stdout    bytes.Buffer
	stderr    bytes.Buffer
	startTime time.Time
	endTime   time.Time
	timeout   time.Duration
	failed    bool
}

func NewCommand(name string, arg ...string) *Command {
	cmd := &Command{
		cmd: exec.Command(name, arg...),
	}

	return cmd
}

func (self *Command) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
}

func (self *Command) Run() error {
	self.cmd.Stdout = &self.stdout
	self.cmd.Stderr = &self.stderr

	if err := self.cmd.Start(); err != nil {
		return err
	}

	self.startTime = time.Now()

	return nil
}

func (self *Command) Wait() error {
	exitCode := 0
	if err := self.doWait(); err != nil {
		self.failed = true

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			if err != errorTimeout {
				return err
			} else {
				exitCode = -1
				self.cmd.Process.Kill()
			}
		}
	}

	self.endTime = time.Now()
	self.buildResponse(exitCode)

	return nil
}

func (self *Command) Kill() error {
	return self.cmd.Process.Kill()
}

func (self *Command) doWait() error {
	if self.timeout != 0 {
		return self.doWaitWithTimeout()
	}

	return self.doWaitWithoutTimeout()
}

func (self *Command) doWaitWithoutTimeout() error {
	return self.cmd.Wait()
}

func (self *Command) doWaitWithTimeout() error {
	done := make(chan error)
	go func() {
		done <- self.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(self.timeout):
		return errorTimeout
	}
}

func (self *Command) buildResponse(exitCode int) {
	response := &ExecutionResponse{
		Stdout:   self.stdout.Bytes(),
		Stderr:   self.stderr.Bytes(),
		ExitCode: exitCode,
		Elapsed:  self.endTime.Sub(self.startTime),
		Pid:      self.cmd.Process.Pid,
		Failed:   self.failed,
	}

	self.response = response
}

func (self *Command) GetResponse() *ExecutionResponse {
	return self.response
}
