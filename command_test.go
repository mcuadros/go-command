package command

import (
	"fmt"
	"testing"
)
import . "launchpad.net/gocheck"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (self *CommandSuite) TestSetDefaultsBasic(c *C) {
	cmd := NewCommand("./test", "-exit=0", "-time=1")

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return
	}

	response := cmd.GetResponse()
	fmt.Println(response)

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
}
