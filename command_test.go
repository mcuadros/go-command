package command

import (
	"fmt"
	"testing"
	"time"
)
import . "launchpad.net/gocheck"

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (self *CommandSuite) TestBasic(c *C) {
	cmd := NewCommand("./test", "-exit=0", "-time=1")
	cmd.Run()
	cmd.Wait()

	response := cmd.GetResponse()

	c.Assert(response.Failed, Equals, false)
	c.Assert(response.ExitCode, Equals, 0)
	c.Assert(response.Stdout, HasLen, 588895)
	c.Assert(response.Stderr, HasLen, 0)
	c.Assert(response.Pid, Not(Equals), 0)
	c.Assert(int(response.Elapsed/time.Second), Aprox, 1, 0.10)
}

var Aprox Checker = &aproxChecker{&CheckerInfo{
	Name: "Aprox",
	Params: []string{
		"Value",
		"Aproximate value",
		"Margin",
	},
}}

type aproxChecker struct {
	*CheckerInfo
}

func (c *aproxChecker) Check(params []interface{}, names []string) (result bool, error string) {
	var value, match, margin float64

	if param, ok := params[0].(int); !ok {
		return false, "First parameter is not an int"
	} else {
		value = float64(param)
	}

	if param, ok := params[1].(int); !ok {
		return false, "Second parameter is not an int"
	} else {
		match = float64(param)
	}

	if param, ok := params[2].(float64); !ok {
		return false, "Third parameter is not an float64"
	} else {
		margin = float64(param)
	}

	lowBound := match - (match * margin)
	if value < lowBound {
		return false, fmt.Sprintf("Lower than bound %f", lowBound)
	}

	highBound := match + (match * margin)
	if value > highBound {
		return false, fmt.Sprintf("Higher than bound %f", highBound)
	}

	return true, ""
}
