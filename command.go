package command

import (
	"bytes"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

var shell = [2]string{"/bin/sh", "-s"}
var errorTimeout = errors.New("error: execution timeout")

type ExecutionResponse struct {
	RealTime time.Duration
	UserTime time.Duration
	SysTime  time.Duration
	Stdout   []byte
	Stderr   []byte
	Pid      int
	ExitCode int
	Failed   bool
	Rusage   *syscall.Rusage
}

type Command struct {
	response  *ExecutionResponse
	cmd       *exec.Cmd
	stdin     *bytes.Buffer
	stdout    bytes.Buffer
	stderr    bytes.Buffer
	startTime time.Time
	endTime   time.Time
	timeout   time.Duration
	failed    bool
}

// NewCommand returns the Command struct to execute the named program with
// the given arguments.
func NewCommand(commands string) *Command {
	cmd := &Command{
		stdin: bytes.NewBufferString(commands),
	}

	return cmd
}

// SetTimeout configure the limit amount of time to run
func (self *Command) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
func (self *Command) Run() error {
	self.cmd = exec.Command(shell[0], shell[1])
	self.cmd.Stdin = self.stdin
	self.cmd.Stdout = &self.stdout
	self.cmd.Stderr = &self.stderr

	if err := self.cmd.Start(); err != nil {
		return err
	}

	self.startTime = time.Now()

	return nil
}

// Wait waits for the Command to exit.
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
				self.Kill()
			}
		}
	}

	self.endTime = time.Now()
	self.buildResponse(exitCode)

	return nil
}

// Kill causes the Command to exit immediately.
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
	go func() {
		time.Sleep(self.timeout)
		self.Kill()
	}()

	return self.cmd.Wait()
}

func (self *Command) buildResponse(exitCode int) {
	response := &ExecutionResponse{
		ExitCode: exitCode,
		RealTime: self.endTime.Sub(self.startTime),
		UserTime: self.cmd.ProcessState.UserTime(),
		SysTime:  self.cmd.ProcessState.UserTime(),
		Rusage:   self.cmd.ProcessState.SysUsage().(*syscall.Rusage),
		Stdout:   self.stdout.Bytes(),
		Stderr:   self.stderr.Bytes(),
		Pid:      self.cmd.Process.Pid,
		Failed:   self.failed,
	}

	self.response = response
}

// GetResponse returns a ExecutionResponse struct, must be called at the end
// of the execution
func (self *Command) GetResponse() *ExecutionResponse {
	return self.response
}
