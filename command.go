package command

import (
	"bytes"
	"errors"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/gobs/args"
)

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
	failed    bool
	exitCode  int
	params    struct {
		user        string
		timeout     time.Duration
		workingDir  string
		environment []string
	}
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
func (c *Command) SetTimeout(timeout time.Duration) {
	c.params.timeout = timeout
}

// SetUsername execute a command as this user (need sudo privilegies)
func (c *Command) SetUser(username string) {
	c.params.user = username
}

// SetWorkingDir sets the working directory of the command. If not is set,
// Run runs the command in the calling process's current directory.
func (c *Command) SetWorkingDir(workingDir string) {
	c.params.workingDir = workingDir
}

// SetEnvironment sets the environment of the process. If not is set,
// Run uses the current process's environment.
func (c *Command) SetEnvironment(environment []string) {
	c.params.environment = environment
}

// Run starts the specified command and waits for it to complete.
//
// The returned error is nil if the command runs, has no problems
// copying stdin, stdout, and stderr, and exits with a zero exit
// status.
func (c *Command) Run() error {
	c.buildExecCmd()
	c.setOutput()
	if err := c.setCredentials(); err != nil {
		return err
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.startTime = time.Now()

	return nil
}

func (c *Command) buildExecCmd() {
	arguments := args.GetArgs(c.command)
	aname, err := exec.LookPath(arguments[0])
	if err != nil {
		aname = arguments[0]
	}

	c.cmd = &exec.Cmd{
		Path: aname,
		Args: arguments,
	}

	if c.params.workingDir != "" {
		c.cmd.Dir = c.params.workingDir
	}

	if c.params.environment != nil {
		c.cmd.Env = c.params.environment
	}
}

func (c *Command) setOutput() {
	c.cmd.Stdout = &c.stdout
	c.cmd.Stderr = &c.stderr
}

func (c *Command) setCredentials() error {
	if c.params.user == "" {
		return nil
	}

	uid, gid, err := c.getUidAndGidInfo(c.params.user)
	if err != nil {
		return err
	}

	c.cmd.SysProcAttr = &syscall.SysProcAttr{}
	c.cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}

	return nil
}

func (c *Command) getUidAndGidInfo(username string) (uint32, uint32, error) {
	user, err := user.Lookup(username)
	if err != nil {
		return 0, 0, err
	}

	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)

	return uint32(uid), uint32(gid), nil
}

// Wait waits for the Command to exit.
func (c *Command) Wait() error {
	if err := c.doWait(); err != nil {
		c.failed = true

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				c.exitCode = status.ExitStatus()
			}
		} else {
			if err != errorTimeout {
				return err
			} else {
				c.Kill()
			}
		}
	}

	c.endTime = time.Now()
	c.buildResponse()

	return nil
}

// Kill causes the Command to exit immediately.
func (c *Command) Kill() error {
	c.failed = true
	c.exitCode = -1

	return c.cmd.Process.Kill()
}

func (c *Command) doWait() error {
	if c.params.timeout != 0 {
		return c.doWaitWithTimeout()
	}

	return c.doWaitWithoutTimeout()
}

func (c *Command) doWaitWithoutTimeout() error {
	return c.cmd.Wait()
}

func (c *Command) doWaitWithTimeout() error {
	go func() {
		time.Sleep(c.params.timeout)
		c.Kill()
	}()

	return c.cmd.Wait()
}

func (c *Command) buildResponse() {
	response := &ExecutionResponse{
		RealTime: c.endTime.Sub(c.startTime),
		UserTime: c.cmd.ProcessState.UserTime(),
		SysTime:  c.cmd.ProcessState.UserTime(),
		Rusage:   c.cmd.ProcessState.SysUsage().(*syscall.Rusage),
		Stdout:   c.stdout.Bytes(),
		Stderr:   c.stderr.Bytes(),
		Pid:      c.cmd.Process.Pid,
		Failed:   c.failed,
		ExitCode: c.exitCode,
		User:     c.params.user,
	}

	c.response = response
}

// GetResponse returns a ExecutionResponse struct, must be called at the end
// of the execution
func (c *Command) GetResponse() *ExecutionResponse {
	return c.response
}
