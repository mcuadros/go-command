go-command [![Build Status](https://travis-ci.org/mcuadros/go-command.png?branch=master)](https://travis-ci.org/mcuadros/go-command) [![GoDoc](https://godoc.org/github.com/mcuadros/go-command?status.png)](http://godoc.org/github.com/mcuadros/go-command)
==============================

Yet another wrapper of os/exec

Installation
------------

The recommended way to install go-command

```
go get github.com/mcuadros/go-command
```

Examples
--------

How import the package

```go
import (
    "github.com/mcuadros/go-command"
)
```

Basic example:

```go
cmd := NewCommand("./test", "-exit=0", "-time=1", "-max=1")
if err := cmd.Run(); err != nil {
    panic(err)
}

if err := cmd.Wait(); err != nil {
    panic(err)
}

response := cmd.GetResponse()
fmt.Printf("Failed: %v, ExitCode: %d, Elapsed: %s",
    response.Failed,
    response.ExitCode,
    response.RealTime,
)

//Return: Failed: false, ExitCode: 0, RealTime: 1.004233166s
```

License
-------

MIT, see [LICENSE](LICENSE)
