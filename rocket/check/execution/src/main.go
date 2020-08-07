package main

import (
	"hello/pkg/hello"
	"os"
)

func main() {

	hello.Inputs{
		Username:    os.Getenv("USERNAME"),
		Password:    os.Getenv("PASSWORD"),
		ExecutionID: os.Getenv("EXECUTION_ID"),
		Context:     os.Getenv("CONTEXT"),
	}.Run()
}
