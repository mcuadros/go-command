package command

import (
	"flag"
	"fmt"
	"os"
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
	cmd := NewCommand("./tests/test -exit=0 -time=1 -max=100000")
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
	cmd := NewCommand("./tests/test -exit=0 -time=2")
	cmd.SetTimeout(1 * time.Second)
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, true)
	c.Assert(response.ExitCode, Equals, -1)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
}

func (self *CommandSuite) TestKill(c *C) {
	cmd := NewCommand("./tests/test -exit=0 -time=2")
	cmd.Run()

	go func() {
		time.Sleep(1 * time.Second)
		cmd.Kill()
	}()

	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, true)
	c.Assert(response.ExitCode, Equals, -1)
	c.Assert(int(response.RealTime/time.Second), Equals, 1)
	c.Assert(int(response.UserTime), Not(Equals), 0)
}

func (self *CommandSuite) TestSetUser(c *C) {
	if *travis {
		c.Skip("Running at TravisCI")
	}

	cmd := NewCommand("./tests/test -exit=0 -time=1")
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
}

func (self *CommandSuite) TestSetWorkingDir(c *C) {
	cmd := NewCommand("./test -exit=0 -wd")

	cwd, _ := os.Getwd()
	wd := cwd + "/tests"

	cmd.SetWorkingDir(wd)
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
	c.Assert(string(response.Stdout), Equals, wd+"\n")
}

func (self *CommandSuite) TestSetEnvironment(c *C) {
	cmd := NewCommand("./tests/test -exit=0 -env")
	cmd.SetEnvironment([]string{"FOO=bar"})
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
	c.Assert(string(response.Stdout), Equals, "FOO=bar\n")
}
