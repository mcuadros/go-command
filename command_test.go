package command

import (
	"flag"
	"fmt"
	"testing"
	"time"
)

import . "launchpad.net/gocheck"

var travis = flag.Bool("travis", false, "Enable it if the tests runs in TravisCI")

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (self *CommandSuite) TestBasic(c *C) {
	cmd := NewCommand("./test -exit=0 -time=1")
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
	c.Assert(response.Stdout, HasLen, 588895)
	c.Assert(response.Stderr, HasLen, 0)
	c.Assert(response.Pid, Not(Equals), 0)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
	c.Assert(int(response.SysTime), Not(Equals), 0)
	c.Assert(int(response.Rusage.Utime.Usec), Not(Equals), 0)
}

func (self *CommandSuite) TestBasicWithTimeout(c *C) {
	cmd := NewCommand("./test -exit=0 -time=2")
	cmd.SetTimeout(1 * time.Second)
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, true)
	c.Assert(response.ExitCode, Equals, -1)
	c.Assert(response.Stdout, HasLen, 588895)
	c.Assert(response.Stderr, HasLen, 0)
	c.Assert(response.Pid, Not(Equals), 0)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
}

func (self *CommandSuite) TestKill(c *C) {
	cmd := NewCommand("./test -exit=0 -time=2")

	go func() {
		time.Sleep(1 * time.Second)
		cmd.Kill()
	}()

	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, true)
	c.Assert(response.ExitCode, Equals, -1)
	c.Assert(response.Stdout, HasLen, 588895)
	c.Assert(response.Stderr, HasLen, 0)
	c.Assert(response.Pid, Not(Equals), 0)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
}

func (self *CommandSuite) TestSetUser(c *C) {
	if *travis {
		c.Skip("Running at TravisCI")
	}

	cmd := NewCommand("./test -exit=0 -time=1")
	cmd.SetUser("daemon")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		c.Fail()
		return
	}

	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
	c.Assert(response.Stdout, HasLen, 588895)
	c.Assert(response.Stderr, HasLen, 0)
	c.Assert(response.Pid, Not(Equals), 0)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
	c.Assert(int(response.SysTime), Not(Equals), 0)
	c.Assert(int(response.Rusage.Utime.Usec), Not(Equals), 0)
}
