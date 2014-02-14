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
var maxFlag = flag.Int("max", 0, "Max count")
var workingDir = flag.Bool("wd", false, "Print working dir")
var enviroment = flag.Bool("env", false, "Print enviroment vars")

func main() {
	flag.Parse()

	if *workingDir {
		printWorkingDir()
	}

	if *enviroment {
		printEnviroment()
	}

	if *maxFlag != 0 {
		go printCount(*maxFlag)
	}

	time.Sleep(time.Duration(*timeFlag) * time.Second)
	os.Exit(*exitFlag)
}

func printCount(count int) {
	i := 0
	for {
		i++
		fmt.Println(i)

		if i >= count {
			break
		}
	}
}

func printWorkingDir() {
	pwd, _ := os.Getwd()
	fmt.Println(pwd)
}

func printEnviroment() {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}
}
