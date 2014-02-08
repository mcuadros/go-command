package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// note, that variables are pointers
var timeFlag = flag.Int("time", 1, "Execution time")
var exitFlag = flag.Int("exit", 0, "Exit Code")
var maxFlag = flag.Int("max", 100000, "Max count")

func main() {
	flag.Parse()

	i := 0
	go func() {
		for {
			i++
			fmt.Println(i)

			if i >= *maxFlag {
				break
			}
		}
	}()

	time.Sleep(time.Duration(*timeFlag) * time.Second)
	os.Exit(*exitFlag)
}
