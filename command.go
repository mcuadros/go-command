package command

import (
	"bytes"
	"errors"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

import "github.com/gobs/args"

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
	User     string
	Rusage   *syscall.Rusage
}

type Command struct {
	response  *ExecutionResponse
	cmd       *exec.Cmd
	command   string
	stdout    bytes.Buffer
	stderr    bytes.Buffer
	startTime time.Time
	endTime   time.Time
	timeout   time.Duration
	failed    bool
	exitCode  int
	user      string
}

// NewCommand returns the Command struct to execute the named program with
// the given arguments.
func NewCommand(command string) *Command {
	cmd := &Command{
		command: command,
	}

	return cmd
}

// SetTimeout configure the limit amount of time to run
func (self *Command) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
}

// SetUsername execute a command as this user (need sudo privilegies)
func (self *Command) SetUser(username string) {
	self.user = username
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
func (self *Command) Run() error {
	self.buildExecCmd()
	self.setOutput()
	if err := self.setCredentials(); err != nil {
		return err
	}

	if err := self.cmd.Start(); err != nil {
		return err
	}

	self.startTime = time.Now()

	return nil
}

func (self *Command) buildExecCmd() {
	arguments := args.GetArgs(self.command)
	aname, err := exec.LookPath(arguments[0])
	if err != nil {
		aname = arguments[0]
	}

	self.cmd = &exec.Cmd{
		Path: aname,
		Args: arguments,
	}
}

func (self *Command) setOutput() {
	self.cmd.Stdout = &self.stdout
	self.cmd.Stderr = &self.stderr
}

func (self *Command) setCredentials() error {
	if self.user == "" {
		return nil
	}

	uid, gid, err := self.getUidAndGidInfo(self.user)
	if err != nil {
		return err
	}

	self.cmd.SysProcAttr = &syscall.SysProcAttr{}
	self.cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	return nil
}

func (self *Command) getUidAndGidInfo(username string) (uint32, uint32, error) {
	user, err := user.Lookup(username)
	if err != nil {
		return 0, 0, err
	}

	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)

	return uint32(uid), uint32(gid), nil
}

// Wait waits for the Command to exit.
func (self *Command) Wait() error {
	if err := self.doWait(); err != nil {
		self.failed = true

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				self.exitCode = status.ExitStatus()
			}
		} else {
			if err != errorTimeout {
				return err
			} else {
				self.Kill()
			}
		}
	}

	self.endTime = time.Now()
	self.buildResponse()

	return nil
}

// Kill causes the Command to exit immediately.
func (self *Command) Kill() error {
	self.failed = true
	self.exitCode = -1

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

func (self *Command) buildResponse() {
	response := &ExecutionResponse{
		RealTime: self.endTime.Sub(self.startTime),
		UserTime: self.cmd.ProcessState.UserTime(),
		SysTime:  self.cmd.ProcessState.UserTime(),
		Rusage:   self.cmd.ProcessState.SysUsage().(*syscall.Rusage),
		Stdout:   self.stdout.Bytes(),
		Stderr:   self.stderr.Bytes(),
		Pid:      self.cmd.Process.Pid,
		Failed:   self.failed,
		ExitCode: self.exitCode,
		User:     self.user,
	}

	self.response = response
}

// GetResponse returns a ExecutionResponse struct, must be called at the end
// of the execution
func (self *Command) GetResponse() *ExecutionResponse {
	return self.response
}
