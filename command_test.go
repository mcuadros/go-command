package command

import (
	"testing"
)
import . "launchpad.net/gocheck"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (self *CommandSuite) TestSetDefaultsBasic(c *C) {
	cmd := NewCommand("./test", "-exit=0", "-time=1")
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
}
